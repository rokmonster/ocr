package opencvutils

import (
	"image"

	"github.com/sirupsen/logrus"
)

func OpenCVFindCoordsWithDebug(img, search image.Image, debug bool) (int, int, int, int, float32) {
	logrus.Errorf("Finding coordinates of image inside image is not implemented in build without opencv")

	// TODO: Not implemented
	return 0, 0, 0, 0, 0
}
