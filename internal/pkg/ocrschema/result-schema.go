package ocrschema

type OCRResult struct {
	Filename string                 `json:"filename"`
	Data     map[string]interface{} `json:"data"`
}
