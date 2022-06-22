package scannerconfig

import (
	"flag"
	"os"

	"github.com/rokmonster/ocr/internal/pkg/config"
)

type ROKOCRConfiguration struct {
	config.CommonConfiguration
	ForceTemplate string
}

func Parse() ROKOCRConfiguration {
	var flags ROKOCRConfiguration

	flag.StringVar(&flags.MediaDirectory, "media", "./media", "folder where all files to scan is placed")
	flag.StringVar(&flags.TemplatesDirectory, "templates", "./templates", "templates dir")
	flag.StringVar(&flags.TessdataDirectory, "tessdata", "./tessdata", "tesseract data files directory")
	flag.StringVar(&flags.OutputDirectory, "output", "./out", "output dir")
	flag.StringVar(&flags.TmpDirectory, "tmp", os.TempDir(), "Directory for temporary files (cropped ones)")
	flag.StringVar(&flags.ForceTemplate, "forceTemplate", "", "Force a specific template")
	flag.Parse()

	return flags
}
