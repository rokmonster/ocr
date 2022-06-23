package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/rokmonster/ocr/internal/pkg/imgutils"
	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

func openCVFindCoords(img, search image.Image) (int, int) {
	mat, _ := gocv.ImageToMatRGB(img)
	tpl, _ := gocv.ImageToMatRGB(search)

	result := gocv.NewMat()

	gocv.MatchTemplate(mat, tpl, &result, gocv.TmCcoeffNormed, gocv.NewMat())

	newResult := gocv.NewMat()
	gocv.Normalize(result, &newResult, 0, 1, gocv.NormMinMax)

	minConfidence, maxConfidence, minLoc, maxLoc := gocv.MinMaxLoc(newResult)
	fmt.Println(minConfidence, maxConfidence, minLoc, maxLoc)

	gocv.Rectangle(&mat, image.Rect(maxLoc.X, maxLoc.Y, maxLoc.X+tpl.Size()[0], maxLoc.Y+tpl.Size()[1]), color.RGBA{255, 0, 0, 1}, 10)

	gocv.IMWrite("found.jpg", mat)

	size := tpl.Size()
	return maxLoc.X + (size[0] / 2), maxLoc.Y + (size[1] / 2)
}

func main() {
	original, _ := imgutils.ReadImageFile("./out.png")
	findMe, _ := imgutils.ReadImageFile("./web/static/cuts/close.png")

	x, y := openCVFindCoords(original, findMe)
	logrus.Infof("TapPoint: %v, %v", x, y)
}
