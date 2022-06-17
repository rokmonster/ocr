package ocrschema

type OCRResponse struct {
	Filename string                 `json:"filename"`
	Data     map[string]interface{} `json:"data"`
}
