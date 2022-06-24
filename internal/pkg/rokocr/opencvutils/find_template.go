package opencvutils

import (
	"image"
	"image/color"

	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

func OpenCVFindCoords(img, search image.Image) (int, int) {
	return OpenCVFindCoordsWithDebug(img, search, false)
}

func OpenCVFindCoordsWithDebug(img, search image.Image, debug bool) (int, int) {
	mat, _ := gocv.ImageToMatRGB(img)
	tpl, _ := gocv.ImageToMatRGB(search)

	result := gocv.NewMat()

	gocv.MatchTemplate(mat, tpl, &result, gocv.TmCcoeffNormed, gocv.NewMat())

	newResult := gocv.NewMat()
	gocv.Normalize(result, &newResult, 0, 1, gocv.NormMinMax)

	minConfidence, maxConfidence, minLoc, maxLoc := gocv.MinMaxLoc(newResult)

	size := tpl.Size()

	if debug {
		logrus.Infof("Min: %.3f, Max: %.3f, minLoc: %v, maxLoc: %v", minConfidence, maxConfidence, minLoc, maxLoc)
		gocv.Rectangle(&mat, image.Rect(
			maxLoc.X,
			maxLoc.Y,
			maxLoc.X+size[0],
			maxLoc.Y+size[1],
		), color.RGBA{R: 255, G: 0, B: 0, A: 1}, 5)

		gocv.IMWrite("./debug.png", mat)
	}

	return maxLoc.X + (size[0] / 2), maxLoc.Y + (size[1] / 2)
}
