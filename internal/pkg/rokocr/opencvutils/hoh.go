package opencvutils

import (
	"fmt"
	"image"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/rokmonster/ocr/internal/pkg/ocrschema"
	"github.com/rokmonster/ocr/internal/pkg/rokocr/tesseractutils"
	"github.com/rokmonster/ocr/internal/pkg/utils"
	"github.com/rokmonster/ocr/internal/pkg/utils/imgutils"
	"github.com/rokmonster/ocr/web"
	"github.com/sirupsen/logrus"
)

type Coordinates struct {
	X, Y, W, H int
}

type HOHResult struct {
	T4         int               `json:"t4"`
	T5         int               `json:"t5"`
	RawResults map[string]string `json:"_raw"`
}

func HOHScan(original image.Image, tessdataDir string) HOHResult {
	locations := make(map[string]Coordinates, 0)

	files := []string{}
	entries, _ := web.RecognitionFS.ReadDir("recognition/hoh")
	for _, f := range entries {
		files = append(files, fmt.Sprintf("recognition/hoh/%v", f.Name()))
	}

	for _, fname := range files {
		f, e := web.RecognitionFS.Open(fname)
		utils.Panic(e)

		findMe, _ := imgutils.ReadImage(f)
		x, y, w, h, c := OpenCVFindCoordsWithDebug(original, findMe, true)

		if c > 0.99 {
			logrus.Infof("Location [%v]: %v, %v, %v, %v (conf: %v)", filepath.Base(fname), x, y, w, h, c)
			locations[filepath.Base(fname)] = Coordinates{x, y, w, h}
		} else {
			logrus.Warnf("Location [%v]: %v, %v, %v, %v (conf: %v) -- ignoring", filepath.Base(fname), x, y, w, h, c)
		}
	}

	distanceX := findAverageDistanceX(locations)
	width := findAverageWidth(locations)

	if distanceX < 0 {
		return HOHResult{}
	}

	DeadT4 := 0
	DeadT5 := 0

	rawResults := map[string]string{}

	for n, l := range locations {
		imgNew, _ := imgutils.CropImage(original, image.Rect(l.X+width, l.Y, l.X+distanceX, l.Y+l.H))

		imgutils.WritePNGImage(imgNew, n)
		text, _ := tesseractutils.ParseText(n, ocrschema.OCRSchema{
			Crop: &ocrschema.OCRCrop{
				X: 0,
				Y: 0,
				W: imgNew.Bounds().Dx(),
				H: imgNew.Bounds().Dy(),
			},
			OEM:       1,
			PSM:       7,
			Languages: []string{"eng"},
			AllowList: []interface{}{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", ","},
		}, tessdataDir)

		value, _ := strconv.Atoi(strings.ReplaceAll(text, ",", ""))

		rawResults[n] = text

		if strings.Contains(n, "_t5_") {
			DeadT5 = DeadT5 + value
		} else if strings.Contains(n, "_t4_") {
			DeadT4 = DeadT4 + value
		}
		// _ = os.Remove(n)
	}

	return HOHResult{T4: DeadT4, T5: DeadT5, RawResults: rawResults}
}

func findAverageWidth(locations map[string]Coordinates) int {
	sum := 0
	for _, l := range locations {
		sum = sum + l.W
	}
	return sum / len(locations)
}

func findAverageDistanceX(locations map[string]Coordinates) int {
	var values []int
	for _, l := range locations {
		values = append(values, l.X)
	}
	sort.Ints(values)

	var diffs []int
	for i := 0; i < len(values)-1; i++ {
		diff := values[i+1] - values[i]
		if diff > 10 {
			diffs = append(diffs, diff)
		}
	}

	sum := 0
	for _, x := range diffs {
		sum = sum + x
	}

	if len(diffs) == 0 {
		return -1
	}

	return sum / len(diffs)
}
