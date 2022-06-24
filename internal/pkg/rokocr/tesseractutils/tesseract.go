package tesseractutils

import (
	"fmt"
	"strings"

	"github.com/otiai10/gosseract/v2"
	schema "github.com/rokmonster/ocr/internal/pkg/ocrschema"
	log "github.com/sirupsen/logrus"
)

func ParseText(imageFileName string, schema schema.OCRSchema, tessdata string) (string, error) {
	client := gosseract.NewClient()

	_ = client.SetTessdataPrefix(tessdata)
	if len(schema.Languages) > 0 {
		_ = client.SetLanguage(schema.Languages...)
	} else {
		_ = client.SetLanguage("eng")
	}
	_ = client.SetPageSegMode(gosseract.PageSegMode(schema.PSM))

	defer client.Close()

	_ = client.SetImage(imageFileName)

	if len(schema.AllowList) > 0 {
		var whitelistedCharacters []string

		for _, x := range schema.AllowList {
			whitelistedCharacters = append(whitelistedCharacters, fmt.Sprintf("%v", x))
		}

		whitelist := strings.Join(whitelistedCharacters, "")
		_ = client.SetWhitelist(whitelist)
	}

	text, err := client.Text()
	if err != nil {
		log.Fatalf("Error: %s", err)
		return "", err
	}
	return text, nil
}
