package main

import (
	"github.com/rokmonster/ocr/internal/pkg/rokocr/opencvutils"
	"github.com/rokmonster/ocr/internal/pkg/utils"
	"github.com/rokmonster/ocr/internal/pkg/utils/imgutils"
	"github.com/rokmonster/ocr/web"
	"github.com/sirupsen/logrus"
)

func main() {
	original, _ := imgutils.ReadImageFile("./out.png")
	f, e := web.RecognitionFS.Open("recognition/close.png")
	utils.Panic(e)

	findMe, _ := imgutils.ReadImage(f)

	x, y := opencvutils.OpenCVFindCoords(original, findMe)
	logrus.Infof("TapPoint: %v, %v", x, y)
}
