package rokocr

import (
	"fmt"
	"strings"

	"github.com/otiai10/gosseract/v2"
	log "github.com/sirupsen/logrus"
	schema "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/ocrschema"
)

func ParseText(imageFileName string, schema schema.ROKOCRSchema, tessdata string) (string, error) {
	client := gosseract.NewClient()

	client.SetTessdataPrefix(tessdata)
	if len(schema.Languages) > 0 {
		client.SetLanguage(schema.Languages...)
	} else {
		client.SetLanguage("eng")
	}
	client.SetPageSegMode(gosseract.PageSegMode(schema.PSM))

	defer client.Close()

	client.SetImage(imageFileName)

	if len(schema.AllowList) > 0 {
		allowlistString := []string{}

		for _, x := range schema.AllowList {
			allowlistString = append(allowlistString, fmt.Sprintf("%v", x))
		}

		whitelist := strings.Join(allowlistString, "")
		client.SetWhitelist(whitelist)
	}

	text, err := client.Text()
	if err != nil {
		log.Fatalf("Error: %s", err)
		return "", err
	}
	return text, nil
}
