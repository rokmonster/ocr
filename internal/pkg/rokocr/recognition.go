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

func RunRecognition(mediaDir, tessData string, template *schema.RokOCRTemplate) []schema.OCRResponse {
	// scan all the images
	data := []schema.OCRResponse{}
	for _, f := range fileutils.GetFilesInDirectory(mediaDir) {
		result, err := ParseSingleFile(f, tessData, template)
		if err != nil {
			log.Warnf("[%s] %v", filepath.Base(f), err)
			continue
		}
		data = append(data, *result)
	}

	return data
}

func ParseSingleFile(f, tessData string, template *schema.RokOCRTemplate) (*schema.OCRResponse, error) {
	img, err := imgutils.ReadImage(f)
	if err != nil {
		return nil, fmt.Errorf("cant read file: %v", err)
	}

	if template.Matches(img) {
		result := ParseImage(f, img, template, os.TempDir(), tessData)
		return &result, nil
	}

	return nil, fmt.Errorf("image doesn't match the template")
}
