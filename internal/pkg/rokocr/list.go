package rokocr

import (
	"path/filepath"
	"strings"

	"github.com/rokmonster/ocr/internal/pkg/utils/fileutils"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func AvailableLanguages(dir string) []string {
	base, _ := filepath.Abs(dir)

	// always keep english, and always at first index
	languages := []string{"eng"}

	for _, f := range fileutils.GetFilesInDirectory(base) {
		if filepath.Ext(f) == ".traineddata" {
			langcode := strings.TrimSuffix(filepath.Base(f), ".traineddata")
			if !contains(languages, langcode) {
				languages = append(languages, langcode)
			}
		}
	}

	return languages
}
