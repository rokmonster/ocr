package opencvutils

import (
	"image"
)

func OpenCVFindCenterCoords(img, search image.Image) (int, int) {
	return OpenCVFindCenterCoordsWithDebug(img, search, false)
}

func OpenCVFindCenterCoordsWithDebug(img, search image.Image, debug bool) (int, int) {
	x, y, w, h, _ := OpenCVFindCoordsWithDebug(img, search, debug)
	return x + (w / 2), y + (h / 2)
}
