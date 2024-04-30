package tesseractutils

import (
	"image"
	"os"
	"path/filepath"
	"time"

	imgutils2 "github.com/rokmonster/ocr/internal/pkg/utils/imgutils"
	"github.com/rokmonster/ocr/internal/pkg/utils/stringutils"

	log "github.com/sirupsen/logrus"

	schema "github.com/rokmonster/ocr/internal/pkg/ocrschema"
)

func ParseImage(name string, img image.Image, template schema.OCRTemplate, tmpdir, tessdata string) schema.OCRResult {
	log.Debugf("[%s] Processing with template: %s", filepath.Base(name), template.Title)
	start := time.Now()

	results := make(map[string]interface{})

	if template.Width != img.Bounds().Dx() || template.Height != img.Bounds().Dy() {
		log.Debugf("[%s] Need to resize: Original -> %v,%v, Template -> %v, %v", filepath.Base(name), img.Bounds().Dx(), img.Bounds().Dy(), template.Width, template.Height)
		img = imgutils2.ResizeImage(img, template.Width, template.Height)
	}

	for n, s := range template.OCRSchema {
		imgNew, _ := imgutils2.CropImage(img, image.Rect(s.Crop.X, s.Crop.Y, s.Crop.X+s.Crop.W, s.Crop.Y+s.Crop.H))
		croppedName := filepath.Join(tmpdir, n+"_"+stringutils.Random(12)+"_"+filepath.Base(name))
		imgutils2.WritePNGImage(imgNew, croppedName)
		text, _ := ParseText(croppedName, s, tessdata)
		_ = os.Remove(croppedName) // delete the temp file
		log.Debugf("[%s] Extracted '%s' => %v", filepath.Base(name), n, text)
		results[n] = text
	}

	return schema.OCRResult{
		Filename: filepath.Base(name),
		Data:     results,
		Took:     time.Since(start),
	}
}
