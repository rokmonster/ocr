package automatorconfig

import (
	"flag"
	"os"

	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/config"
	adb "github.com/zach-klippenstein/goadb"
)

type AutomatorConfig struct {
	config.CommonConfiguration
	ADBPort int
}

func Parse() AutomatorConfig {
	var flags AutomatorConfig

	flag.StringVar(&flags.MediaDirectory, "media", "./media", "folder where all files to scan is placed")
	flag.StringVar(&flags.TemplatesDirectory, "templates", "./templates", "templates dir")
	flag.StringVar(&flags.TessdataDirectory, "tessdata", "./tessdata", "tesseract data files directory")
	flag.StringVar(&flags.OutputDirectory, "output", "./out", "output dir")
	flag.StringVar(&flags.TmpDirectory, "tmp", os.TempDir(), "Directory for temporary files (cropped ones)")
	flag.IntVar(&flags.ADBPort, "adb-port", adb.AdbPort, "ADB Port")
	flag.Parse()
	return flags
}