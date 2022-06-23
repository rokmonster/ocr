package rokremoteconfig

import (
	"flag"
	"os"

	"github.com/rokmonster/ocr/internal/pkg/config"
	adb "github.com/zach-klippenstein/goadb"
)

type ROKRemoteConfig struct {
	config.CommonConfiguration
	ADBPort int
	Server  string
}

func Parse() ROKRemoteConfig {
	var flags ROKRemoteConfig

	flag.StringVar(&flags.MediaDirectory, "media", "./media", "folder where all files to scan is placed")
	flag.StringVar(&flags.TemplatesDirectory, "templates", "./templates", "templates dir")
	flag.StringVar(&flags.TessdataDirectory, "tessdata", "./tessdata", "tesseract data files directory")
	flag.StringVar(&flags.OutputDirectory, "output", "./out", "output dir")
	flag.StringVar(&flags.TmpDirectory, "tmp", os.TempDir(), "Directory for temporary files (cropped ones)")
	flag.IntVar(&flags.ADBPort, "adb-port", adb.AdbPort, "ADB Port")
	flag.StringVar(&flags.Server, "rok-server", "http://localhost:8080", "rokserver to connect to")
	flag.Parse()
	return flags
}
