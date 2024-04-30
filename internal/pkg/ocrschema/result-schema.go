package ocrschema

import "time"

type OCRResult struct {
	Filename string                 `json:"filename"`
	Data     map[string]interface{} `json:"data"`
	Took     time.Duration          `json:"duration"`
}
