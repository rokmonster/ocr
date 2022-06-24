package ocrschema

import (
	"github.com/rokmonster/ocr/internal/pkg/utils/fileutils"
	"github.com/rokmonster/ocr/internal/pkg/utils/imgutils"
	"path/filepath"

	"github.com/corona10/goimagehash"
	log "github.com/sirupsen/logrus"
)

func LoadTemplates(directory string) []OCRTemplate {
	var templates []OCRTemplate
	for _, f := range fileutils.GetFilesInDirectory(directory) {
		if filepath.Ext(f) == ".json" {
			template, err := LoadTemplate(f)
			if err == nil {
				log.Debugf("Loaded template: %s => %s, hash: %s", f, template.Title, template.Fingerprint)
				templates = append(templates, template)
			} else {
				log.Errorf("Failed to load template: %v => %v", filepath.Base(f), err)
			}
		}
	}
	return templates
}

func FindTemplate(mediaDir string, availableTemplate []OCRTemplate) OCRTemplate {
	for _, file := range fileutils.GetFilesInDirectory(mediaDir) {
		img, err := imgutils.ReadImageFile(file)
		if err != nil {
			log.Debugf("[%s] => error: %v", filepath.Base(file), err)
			continue
		}

		imagehash, _ := goimagehash.DifferenceHash(img)
		template := PickTemplate(imagehash, availableTemplate)
		return template
	}
	// pick first template if no images found?
	return availableTemplate[0]
}

func PickTemplate(hash *goimagehash.ImageHash, availableTemplate []OCRTemplate) OCRTemplate {
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
