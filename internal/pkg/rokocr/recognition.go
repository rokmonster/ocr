package rokocr

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/fileutils"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/imgutils"
	schema "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/ocrschema"
)

func RunRecognition(mediaDir, tessData string, template *schema.RokOCRTemplate, force bool) []schema.OCRResponse {
	// scan all the images
	data := []schema.OCRResponse{}
	dir, _ := filepath.Abs(mediaDir)
	files := fileutils.GetFilesInDirectory(dir)
	for i, f := range files {
		log.Infof("[%04d/%04d] Parsing image: %v", i+1, len(files), filepath.Base(f))
		result, err := ParseSingleFile(f, tessData, template, force)
		if err != nil {
			log.Warnf("[%s] %v", filepath.Base(f), err)
			continue
		}
		data = append(data, *result)
	}

	return data
}

func ParseSingleFile(f, tessData string, template *schema.RokOCRTemplate, force bool) (*schema.OCRResponse, error) {
	img, err := imgutils.ReadImage(f)
	if err != nil {
		return nil, fmt.Errorf("cant read file: %v", err)
	}

	if template.Matches(img) || force {
		result := ParseImage(f, img, template, os.TempDir(), tessData)
		return &result, nil
	}

	return nil, fmt.Errorf("image doesn't match the template")
}
