package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"golang.org/x/image/draw"

	"github.com/corona10/goimagehash"
	"github.com/otiai10/gosseract/v2"
)

type RokOCRTemplate struct {
	Title       string                  `json:"title,omitempty"`
	Version     string                  `json:"version,omitempty"`
	Author      string                  `json:"author,omitempty"`
	Width       int                     `json:"width,omitempty"`
	Height      int                     `json:"height,omitempty"`
	OCRSchema   map[string]ROKOCRSchema `json:"ocr_schema,omitempty"`
	Fingerprint string                  `json:"fingerprint,omitempty"`
	Threshold   int                     `json:"threshold,omitempty"`
	Table       []ROKTableField         `json:"table,omitempty"`
}

func (b *RokOCRTemplate) Hash() *goimagehash.ImageHash {
	result, _ := strconv.ParseUint(b.Fingerprint, 16, 64)
	return goimagehash.NewImageHash(uint64(result), goimagehash.DHash)
}

type ROKOCRSchema struct {
	Callback  []string      `json:"callback,omitempty"`
	Languages []string      `json:"lang,omitempty"`
	OEM       int           `json:"oem,omitempty"`
	PSM       int           `json:"psm,omitempty"`
	Crop      OCRCrop       `json:"crop,omitempty"`
	AllowList []interface{} `json:"allowlist,omitempty"`
}

type ROKTableField struct {
	Title string
	Field string
	Bold  bool
	Color string
}

func (b *ROKTableField) UnmarshalJSON(data []byte) error {

	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	b.Title, _ = v[0].(string)
	b.Field, _ = v[1].(string)
	b.Bold = v[2].(bool)
	b.Color = v[3].(string)

	return nil
}

type OCRCrop struct {
	X int
	Y int
	W int
	H int
}

func (b *OCRCrop) UnmarshalJSON(data []byte) error {

	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	b.X = int(v[0].(float64))
	b.Y = int(v[1].(float64))
	b.W = int(v[2].(float64))
	b.H = int(v[3].(float64))

	return nil
}

func readImage(filename string) (image.Image, error) {
	imgfile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer imgfile.Close()

	// try to decode as PNG
	img, err := png.Decode(imgfile)
	if err == nil {
		return img, nil
	}

	// try to decode as PNG
	img, err = jpeg.Decode(imgfile)
	if err == nil {
		return img, nil
	}

	return nil, fmt.Errorf("unsupported file format")
}

func getFilesInDirectory(directory string) []string {
	files := []string{}

	dir, _ := ioutil.ReadDir(directory)

	for _, file := range dir {
		if !file.IsDir() {
			files = append(files, directory+"/"+file.Name())
		}
	}

	return files
}

type ROKOCRConfig struct {
	MediaDirectory     string
	TemplatesDirectory string
	OutputDirectory    string
	TessdataDirectory  string
	TmpDirectory       string
	DeleteTempFiles    bool
}

var (
	flags ROKOCRConfig
)

func init() {
	flag.StringVar(&flags.MediaDirectory, "media", "./media", "folder where all files to scan is placed")
	flag.StringVar(&flags.TemplatesDirectory, "templates", "./templates", "templates dir")
	flag.StringVar(&flags.TessdataDirectory, "tessdata", "./tessdata", "tesseract data files directory")
	flag.StringVar(&flags.OutputDirectory, "output", "./out", "output dir")
	flag.StringVar(&flags.TmpDirectory, "tmp", "/tmp", "Directory for temporary files (cropped ones)")
	flag.Parse()
}

func parseTemplate(fileName string) RokOCRTemplate {
	var t RokOCRTemplate
	b, _ := ioutil.ReadFile(fileName)
	json.Unmarshal(b, &t)
	return t
}

func pickTemplate(name string, hash *goimagehash.ImageHash, availableTemplate []RokOCRTemplate) RokOCRTemplate {
	best := availableTemplate[0]

	for _, t := range availableTemplate {
		distance, _ := t.Hash().Distance(hash)
		bestDistance, _ := best.Hash().Distance(hash)
		if distance < bestDistance {
			best = t
		}
	}

	return best
}

func resizeImage(src image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

// cropImage takes an image and crops it to the specified rectangle.
func cropImage(img image.Image, crop image.Rectangle) (image.Image, error) {
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	// img is an Image interface. This checks if the underlying value has a
	// method called SubImage. If it does, then we can use SubImage to crop the
	// image.
	simg, ok := img.(subImager)
	if !ok {
		return nil, fmt.Errorf("image does not support cropping")
	}

	return simg.SubImage(crop), nil
}

// writeImage writes an Image back to the disk.
func writeImage(img image.Image, name string) error {
	fd, err := os.Create(name)
	if err != nil {
		log.Errorf("failed to write: %v", err)
		return err
	}
	defer fd.Close()

	return png.Encode(fd, img)
}

func tessaractOCRText(imageFileName string, schema ROKOCRSchema) (string, error) {
	client := gosseract.NewClient()

	client.SetTessdataPrefix(flags.TessdataDirectory)
	if len(schema.Languages) > 0 {
		client.SetLanguage(schema.Languages...)
	} else {
		client.SetLanguage("eng")
	}
	client.SetPageSegMode(gosseract.PageSegMode(schema.PSM))

	defer client.Close()

	client.SetImage(imageFileName)

	if len(schema.AllowList) > 0 {
		allowlistString := []string{}

		for _, x := range schema.AllowList {
			allowlistString = append(allowlistString, fmt.Sprintf("%v", x))
		}

		whitelist := strings.Join(allowlistString, "")
		client.SetWhitelist(whitelist)
	}

	text, err := client.Text()
	if err != nil {
		log.Errorf("Error: %s", err)
		return "", err
	}
	return text, nil
}

func parseImage(name string, img image.Image, template *RokOCRTemplate) OCRResponse {
	log.Infof("[%s] Processing with template: %s", filepath.Base(name), template.Title)

	results := make(map[string]interface{})

	if template.Width != img.Bounds().Dx() || template.Height != img.Bounds().Dy() {
		log.Warnf("[%s] Need to resize: Original -> %v,%v, Template -> %v, %v", filepath.Base(name), img.Bounds().Dx(), img.Bounds().Dy(), template.Width, template.Height)
		img = resizeImage(img, template.Width, template.Height)
	}

	for n, s := range template.OCRSchema {
		imgNew, _ := cropImage(img, image.Rect(s.Crop.X, s.Crop.Y, s.Crop.X+s.Crop.W, s.Crop.Y+s.Crop.H))
		croppedName := flags.TmpDirectory + "/" + n + "_" + filepath.Base(name)
		writeImage(imgNew, croppedName)
		text, _ := tessaractOCRText(croppedName, s)
		_ = os.Remove(croppedName) // delete the temp file
		log.Debugf("[%s] Extracted '%s' => %v", filepath.Base(name), n, text)
		results[n] = text
	}
	return results
}

type OCRResponse map[string]interface{}

func findTemplate(dir string, availableTemplate []RokOCRTemplate) *RokOCRTemplate {
	for _, f := range getFilesInDirectory(flags.MediaDirectory) {
		img, err := readImage(f)
		if err != nil {
			log.Errorf("[%s] => error: %v", filepath.Base(f), err)
			continue
		}

		hash2, _ := goimagehash.DifferenceHash(img)
		template := pickTemplate(f, hash2, availableTemplate)
		return &template
	}
	return nil
}

func printResultsTable(data []OCRResponse, template *RokOCRTemplate) {
	headers := []string{}
	for _, x := range template.Table {
		headers = append(headers, x.Title)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(headers)
	for _, row := range data {
		rowData := []string{}
		for _, x := range template.Table {
			rowData = append(rowData, fmt.Sprintf("%v", row[x.Field]))
		}
		table.Append(rowData)
	}

	table.Render()
}

func writeCSV(data []OCRResponse, template *RokOCRTemplate) {
	headers := []string{}
	for _, x := range template.Table {
		headers = append(headers, x.Title)
	}

	fd, err := os.Create(fmt.Sprintf("%s/%v.csv", flags.OutputDirectory, time.Now().Unix()))
	if err != nil {
		log.Fatalf("Failed to write csv: %v", err)
		return
	}
	defer fd.Close()

	table := csv.NewWriter(fd)
	table.Write(headers)
	for _, row := range data {
		rowData := []string{}
		for _, x := range template.Table {
			rowData = append(rowData, fmt.Sprintf("%v", row[x.Field]))
		}
		table.Write(rowData)
	}
	table.Flush()

}

func loadTemplates(directory string) []RokOCRTemplate {
	templates := []RokOCRTemplate{}
	for _, f := range getFilesInDirectory(directory) {
		template := parseTemplate(f)
		log.Debugf("Loaded template: %s => %s, hash: %s", f, template.Title, template.Fingerprint)
		templates = append(templates, template)
	}
	return templates

}

func runRecognition(directory string, template *RokOCRTemplate) []OCRResponse {
	// scan all the images
	data := []OCRResponse{}
	for _, f := range getFilesInDirectory(flags.MediaDirectory) {
		img, err := readImage(f)
		if err != nil {
			log.Errorf("[%s] => error: %v", filepath.Base(f), err)
			continue
		}

		hash2, _ := goimagehash.DifferenceHash(img)

		distance, _ := hash2.Distance(template.Hash())
		if distance <= template.Threshold {
			result := parseImage(f, img, template)
			data = append(data, result)
		} else {
			log.Warnf("[%s] => hash: %x, distance: %v => SKIPPING", filepath.Base(f), hash2.GetHash(), distance)
		}
	}

	return data
}

func main() {
	templates := loadTemplates(flags.TemplatesDirectory)
	log.Infof("Loaded %v templates", len(templates))

	template := findTemplate(flags.MediaDirectory, templates)
	data := runRecognition(flags.MediaDirectory, template)

	printResultsTable(data, template)
	writeCSV(data, template)
}
