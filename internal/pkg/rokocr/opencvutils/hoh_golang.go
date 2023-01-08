//go:build !opencv

package opencvutils

import (
	"image"

	"github.com/sirupsen/logrus"
)

func HOHScan(original image.Image, tessdataDir string) HOHResult {
	logrus.Errorf("HOHScan is not implemented without opencv")
	return HOHResult{T4: 0, T5: 0, RawResults: map[string]string{}}
}
