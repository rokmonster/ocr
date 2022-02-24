package rokocr

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/config"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/fileutils"
)

func Prepare(flags config.CommonConfiguration) {
	fileutils.Mkdirs(flags.TessdataDirectory)
	fileutils.Mkdirs(flags.MediaDirectory)
	fileutils.Mkdirs(flags.TemplatesDirectory)

	if len(fileutils.GetFilesInDirectory(flags.TessdataDirectory)) == 0 {
		logrus.Warnf("No tesseract trained data found, downloading english & french ones")
		fileutils.Download(filepath.Join(flags.TessdataDirectory, "eng.traineddata"), "https://github.com/tesseract-ocr/tessdata/raw/main/eng.traineddata")
		fileutils.Download(filepath.Join(flags.TessdataDirectory, "fra.traineddata"), "https://github.com/tesseract-ocr/tessdata/raw/main/fra.traineddata")
	}
}
