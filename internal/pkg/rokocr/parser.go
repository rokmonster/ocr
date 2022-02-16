package rokocr

import (
	"image"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/imgutils"
	schema "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/ocrschema"
)

func ParseImage(name string, img image.Image, template *schema.RokOCRTemplate, tmpdir, tessdata string) schema.OCRResponse {
	log.Debugf("[%s] Processing with template: %s", filepath.Base(name), template.Title)

	results := make(map[string]interface{})

	if template.Width != img.Bounds().Dx() || template.Height != img.Bounds().Dy() {
		log.Warnf("[%s] Need to resize: Original -> %v,%v, Template -> %v, %v", filepath.Base(name), img.Bounds().Dx(), img.Bounds().Dy(), template.Width, template.Height)
		img = imgutils.ResizeImage(img, template.Width, template.Height)
	}

	for n, s := range template.OCRSchema {
		imgNew, _ := imgutils.CropImage(img, image.Rect(s.Crop.X, s.Crop.Y, s.Crop.X+s.Crop.W, s.Crop.Y+s.Crop.H))
		croppedName := tmpdir + "/" + n + "_" + filepath.Base(name)
		imgutils.WritePNGImage(imgNew, croppedName)
		text, _ := ParseText(croppedName, s, tessdata)
		_ = os.Remove(croppedName) // delete the temp file
		log.Debugf("[%s] Extracted '%s' => %v", filepath.Base(name), n, text)
		results[n] = text
	}
	return results
}
