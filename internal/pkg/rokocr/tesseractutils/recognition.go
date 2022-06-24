package tesseractutils

import (
	"fmt"
	"github.com/rokmonster/ocr/internal/pkg/utils/fileutils"
	"github.com/rokmonster/ocr/internal/pkg/utils/imgutils"
	"os"
	"path/filepath"

	schema "github.com/rokmonster/ocr/internal/pkg/ocrschema"
	log "github.com/sirupsen/logrus"
)

func RunRecognitionChan(mediaDir, tessData string, template schema.OCRTemplate, force bool) <-chan schema.OCRResult {
	out := make(chan schema.OCRResult)

	go func() {
		dir, _ := filepath.Abs(mediaDir)
		files := fileutils.GetFilesInDirectory(dir)
		for i, f := range files {
			log.Infof("[%04d/%04d] Parsing image: %v", i+1, len(files), filepath.Base(f))
			result, err := ParseSingleFile(f, tessData, template, force)
			if err != nil {
				log.Warnf("[%s] %v", filepath.Base(f), err)
				continue
			}
			out <- *result
		}
		close(out)
	}()

	return out
}

func RunRecognition(mediaDir, tessData string, template schema.OCRTemplate, force bool) []schema.OCRResult {
	var data []schema.OCRResult

	for elem := range RunRecognitionChan(mediaDir, tessData, template, force) {
		data = append(data, elem)
	}

	return data
}

func ParseSingleFile(f, tessData string, template schema.OCRTemplate, force bool) (*schema.OCRResult, error) {
	img, err := imgutils.ReadImageFile(f)
	if err != nil {
		return nil, fmt.Errorf("cant read file: %v", err)
	}

	if template.Matches(img) || force {
		result := ParseImage(f, img, template, os.TempDir(), tessData)
		return &result, nil
	}

	return nil, fmt.Errorf("image doesn't match the template")
}
