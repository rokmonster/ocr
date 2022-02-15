package ocrschema

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/corona10/goimagehash"
	"github.com/sirupsen/logrus"
)

type RokOCRTemplate struct {
	Title       string                  `json:"title,omitempty"`
	Version     string                  `json:"version,omitempty"`
	Author      string                  `json:"author,omitempty"`
	Width       int                     `json:"width,omitempty"`
	Height      int                     `json:"height,omitempty"`
	OCRSchema   map[string]ROKOCRSchema `json:"ocr_schema,omitempty"`
	Fingerprint string                  `json:"fingerprint,omitempty"`
	Threshold   int                     `json:"threshold,omitempty"`
	Table       []ROKTableField         `json:"table,omitempty"`
}

func LoadTemplate(fileName string) RokOCRTemplate {
	var t RokOCRTemplate
	b, _ := ioutil.ReadFile(fileName)
	json.Unmarshal(b, &t)
	return t
}

func (b *RokOCRTemplate) Hash() *goimagehash.ImageHash {
	result, _ := strconv.ParseUint(b.Fingerprint, 16, 64)
	return goimagehash.NewImageHash(uint64(result), goimagehash.DHash)
}

func (b *RokOCRTemplate) Match(hash *goimagehash.ImageHash) bool {
	distance, err := b.Hash().Distance(hash)
	// if we get error, that means this template is no go...
	if err != nil {
		return false
	}

	logrus.Debugf("hash: %x, distance: %v\n", hash.GetHash(), distance)
	return distance <= b.Threshold
}

type ROKOCRSchema struct {
	Callback  []string      `json:"callback,omitempty"`
	Languages []string      `json:"lang,omitempty"`
	OEM       int           `json:"oem,omitempty"`
	PSM       int           `json:"psm,omitempty"`
	Crop      *OCRCrop      `json:"crop,omitempty"`
	AllowList []interface{} `json:"allowlist,omitempty"`
}

func NewNumberField(cropArea *OCRCrop) ROKOCRSchema {
	return ROKOCRSchema{
		Languages: []string{"eng"},
		Callback:  []string{},
		AllowList: []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		PSM:       7,
		OEM:       1,
		Crop:      cropArea,
	}
}

func NewTextField(cropArea *OCRCrop, languages ...string) ROKOCRSchema {
	return ROKOCRSchema{
		Languages: languages,
		Callback:  []string{},
		PSM:       7,
		OEM:       1,
		Crop:      cropArea,
	}
}

type ROKTableField struct {
	Title string
	Field string
	Bold  bool
	Color string
}

func (b *ROKTableField) UnmarshalJSON(data []byte) error {

	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	b.Title, _ = v[0].(string)
	b.Field, _ = v[1].(string)
	b.Bold = v[2].(bool)
	b.Color = v[3].(string)

	return nil
}

type OCRCrop struct {
	X int
	Y int
	W int
	H int
}

func (b *OCRCrop) MarshalJSON() ([]byte, error) {
	return json.Marshal([]int{b.X, b.Y, b.W, b.H})
}

func (b *OCRCrop) UnmarshalJSON(data []byte) error {

	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	b.X = int(v[0].(float64))
	b.Y = int(v[1].(float64))
	b.W = int(v[2].(float64))
	b.H = int(v[3].(float64))

	return nil
}
