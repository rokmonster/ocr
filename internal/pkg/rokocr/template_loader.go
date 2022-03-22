package rokocr

import (
	"path/filepath"

	"github.com/corona10/goimagehash"
	log "github.com/sirupsen/logrus"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/fileutils"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/imgutils"
	schema "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/ocrschema"
)

func LoadTemplates(directory string) []schema.RokOCRTemplate {
	templates := []schema.RokOCRTemplate{}
	for _, f := range fileutils.GetFilesInDirectory(directory) {
		template := schema.LoadTemplate(f)
		log.Debugf("Loaded template: %s => %s, hash: %s", f, template.Title, template.Fingerprint)
		templates = append(templates, *template)
	}
	return templates
}

func FindTemplate(mediaDir string, availableTemplate []schema.RokOCRTemplate) *schema.RokOCRTemplate {
	for _, file := range fileutils.GetFilesInDirectory(mediaDir) {
		img, err := imgutils.ReadImage(file)
		if err != nil {
			log.Debugf("[%s] => error: %v", filepath.Base(file), err)
			continue
		}

		imagehash, _ := goimagehash.DifferenceHash(img)
		template := pickTemplate(imagehash, availableTemplate)
		return &template
	}
	// pick first template if no images found?
	return &availableTemplate[0]
}

func pickTemplate(hash *goimagehash.ImageHash, availableTemplate []schema.RokOCRTemplate) schema.RokOCRTemplate {
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
