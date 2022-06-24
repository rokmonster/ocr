package ocrschema

import (
	"encoding/json"
	"github.com/rokmonster/ocr/internal/pkg/utils/imgutils"
	"image"
	"io/ioutil"
	"strconv"

	"github.com/corona10/goimagehash"
	log "github.com/sirupsen/logrus"
)

type OCRTemplate struct {
	Title       string               `json:"title,omitempty"`
	Version     string               `json:"version,omitempty"`
	Author      string               `json:"author,omitempty"`
	Width       int                  `json:"width,omitempty"`
	Height      int                  `json:"height,omitempty"`
	OCRSchema   map[string]OCRSchema `json:"ocr_schema,omitempty"`
	Fingerprint string               `json:"fingerprint,omitempty"`
	Threshold   int                  `json:"threshold,omitempty"`
	Table       []OCRTableField      `json:"table,omitempty"`
	Checkpoints []OCRCheckpoint      `json:"checkpoints,omitempty"`
}

type OCRCheckpoint struct {
	Crop        *OCRCrop `json:"crop,omitempty"`
	Fingerprint string   `json:"fingerprint,omitempty"`
}

func LoadTemplate(fileName string) (OCRTemplate, error) {
	var t OCRTemplate
	b, _ := ioutil.ReadFile(fileName)
	err := json.Unmarshal(b, &t)
	return t, err
}

func differenceHashFromString(s string) *goimagehash.ImageHash {
	result, _ := strconv.ParseUint(s, 16, 64)
	return goimagehash.NewImageHash(result, goimagehash.DHash)
}

func (b *OCRTemplate) Hash() *goimagehash.ImageHash {
	return differenceHashFromString(b.Fingerprint)
}

func hashMatches(b image.Image, hash *goimagehash.ImageHash) bool {
	imgHash, _ := goimagehash.DifferenceHash(b)
	distance, err := imgHash.Distance(hash)
	// if we get error, that means this template is no go...
	if err != nil {
		return false
	}

	if distance > 0 {
		log.Debugf("Expected hash: %x, real hash: %x, distance: %v", hash.GetHash(), imgHash.GetHash(), distance)
	}

	// max distance allowed here is 1
	return 1 >= distance
}

func (b *OCRTemplate) Matches(img image.Image) bool {
	imageHash, _ := goimagehash.DifferenceHash(img)

	if len(b.Checkpoints) == 0 {
		return b.Match(imageHash)
	}

	// if we have checkpoints, check if all checkpoints matches
	for _, s := range b.Checkpoints {
		expectedHash := differenceHashFromString(s.Fingerprint)
		subImg, _ := imgutils.CropImage(img, s.Crop.CropRectangle())
		if !hashMatches(subImg, expectedHash) {
			log.Debugf("Area %v doesn't match expected hash: %v", s.Crop, s.Fingerprint)
			return false
		}
	}

	return true
}

func (b *OCRTemplate) Match(hash *goimagehash.ImageHash) bool {
	distance, err := b.Hash().Distance(hash)
	// if we get error, that means this template is no go...
	if err != nil {
		return false
	}

	log.Debugf("hash: %x, distance: %v\n", hash.GetHash(), distance)
	return distance <= b.Threshold
}

type OCRSchema struct {
	Callback  interface{}   `json:"callback,omitempty"`
	Languages []string      `json:"lang,omitempty"`
	OEM       int           `json:"oem,omitempty"`
	PSM       int           `json:"psm,omitempty"`
	Crop      *OCRCrop      `json:"crop,omitempty"`
	AllowList []interface{} `json:"allowlist,omitempty"`
}

func NewNumberField(cropArea *OCRCrop) OCRSchema {
	return OCRSchema{
		Languages: []string{"eng"},
		Callback:  []string{},
		AllowList: []interface{}{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		PSM:       7,
		OEM:       1,
		Crop:      cropArea,
	}
}

func NewTextField(cropArea *OCRCrop, languages ...string) OCRSchema {
	return OCRSchema{
		Languages: languages,
		Callback:  []string{},
		PSM:       7,
		OEM:       1,
		Crop:      cropArea,
	}
}

type OCRTableField struct {
	Title string
	Field string
	Bold  bool
	Color string
}

func (b *OCRTableField) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{b.Title, b.Field, b.Bold, b.Color})
}

func (b *OCRTableField) UnmarshalJSON(data []byte) error {

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

func (b *OCRCrop) CropRectangle() image.Rectangle {
	return image.Rect(b.X, b.Y, b.X+b.W, b.Y+b.H)
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
