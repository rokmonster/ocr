package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"

	"github.com/corona10/goimagehash"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/imgutils"
	schema "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/ocrschema"
	rokocr "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/rokocr"
)

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

func parseTemplate(fileName string) schema.RokOCRTemplate {
	var t schema.RokOCRTemplate
	b, _ := ioutil.ReadFile(fileName)
	json.Unmarshal(b, &t)
	return t
}

func pickTemplate(name string, hash *goimagehash.ImageHash, availableTemplate []schema.RokOCRTemplate) schema.RokOCRTemplate {
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

func findTemplate(dir string, availableTemplate []schema.RokOCRTemplate) *schema.RokOCRTemplate {
	for _, f := range getFilesInDirectory(flags.MediaDirectory) {
		img, err := imgutils.ReadImage(f)
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

func printResultsTable(data []schema.OCRResponse, template *schema.RokOCRTemplate) {
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

func writeCSV(data []schema.OCRResponse, template *schema.RokOCRTemplate) {
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

func loadTemplates(directory string) []schema.RokOCRTemplate {
	templates := []schema.RokOCRTemplate{}
	for _, f := range getFilesInDirectory(directory) {
		template := parseTemplate(f)
		log.Debugf("Loaded template: %s => %s, hash: %s", f, template.Title, template.Fingerprint)
		templates = append(templates, template)
	}
	return templates

}

func runRecognition(directory string, template *schema.RokOCRTemplate) []schema.OCRResponse {
	// scan all the images
	data := []schema.OCRResponse{}
	for _, f := range getFilesInDirectory(flags.MediaDirectory) {
		img, err := imgutils.ReadImage(f)
		if err != nil {
			log.Errorf("[%s] => error: %v", filepath.Base(f), err)
			continue
		}

		hash2, _ := goimagehash.DifferenceHash(img)

		distance, _ := hash2.Distance(template.Hash())
		if distance <= template.Threshold {
			result := rokocr.ParseImage(f, img, template, flags.TmpDirectory, flags.TessdataDirectory)
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
