package main

import (
	"github.com/rokmonster/ocr/internal/pkg/rokocr/opencvutils"
	"github.com/rokmonster/ocr/internal/pkg/utils/imgutils"
	"github.com/sirupsen/logrus"
)

func main() {
	original, _ := imgutils.ReadImageFile("./out.png")
	logrus.Printf("%+v", opencvutils.HOHScan(original, "./tessdata"))
}
