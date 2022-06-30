package tesseractutils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rokmonster/ocr/internal/pkg/utils/fileutils"
	"github.com/rokmonster/ocr/internal/pkg/utils/imgutils"
	"github.com/sirupsen/logrus"

	schema "github.com/rokmonster/ocr/internal/pkg/ocrschema"
)

func RunRecognitionChan(mediaDir, tessData string, template schema.OCRTemplate, force bool) <-chan schema.OCRResult {
	out := make(chan schema.OCRResult)

	go func() {
		var wg sync.WaitGroup
		dir, _ := filepath.Abs(mediaDir)
		files := fileutils.GetFilesInDirectory(dir)
		for i, f := range files {
			wg.Add(1)
			go func(index, total int, f string) {
				defer wg.Done()
				start := time.Now()
				result, err := ParseSingleFile(f, tessData, template, force)
				if err != nil {
					logrus.Printf("[%04d/%04d] %v - %v", index, total, filepath.Base(f), err)
					return
				}
				logrus.Printf("[%04d/%04d] %v Took: %v ms", index, total, filepath.Base(f), time.Since(start).Milliseconds())
				out <- *result
			}(i+1, len(files), f)
		}
		wg.Wait()
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
