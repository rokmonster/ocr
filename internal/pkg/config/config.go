package config

import (
	"flag"
	"os"
)

type ROKOCRConfiguration struct {
	MediaDirectory     string
	TemplatesDirectory string
	OutputDirectory    string
	TessdataDirectory  string
	TmpDirectory       string
	DeleteTempFiles    bool
}

func Parse() ROKOCRConfiguration {
	var flags ROKOCRConfiguration

	flag.StringVar(&flags.MediaDirectory, "media", "./media", "folder where all files to scan is placed")
	flag.StringVar(&flags.TemplatesDirectory, "templates", "./templates", "templates dir")
	flag.StringVar(&flags.TessdataDirectory, "tessdata", "./tessdata", "tesseract data files directory")
	flag.StringVar(&flags.OutputDirectory, "output", "./out", "output dir")
	flag.StringVar(&flags.TmpDirectory, "tmp", os.TempDir(), "Directory for temporary files (cropped ones)")
	flag.Parse()

	return flags
}
