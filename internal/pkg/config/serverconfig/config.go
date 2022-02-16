package config

import (
	"flag"
	"os"
)

type RokServerConfiguration struct {
	MediaDirectory     string
	TemplatesDirectory string
	OutputDirectory    string
	TessdataDirectory  string
	TmpDirectory       string
	DeleteTempFiles    bool
	ListenPort         int
	TLS                bool
	TLSDomain          string
}

func Parse() RokServerConfiguration {
	var flags RokServerConfiguration

	flag.StringVar(&flags.MediaDirectory, "media", "./media", "folder where all files to scan is placed")
	flag.StringVar(&flags.TemplatesDirectory, "templates", "./templates", "templates dir")
	flag.StringVar(&flags.TessdataDirectory, "tessdata", "./tessdata", "tesseract data files directory")
	flag.StringVar(&flags.OutputDirectory, "output", "./out", "output dir")
	flag.StringVar(&flags.TmpDirectory, "tmp", os.TempDir(), "Directory for temporary files (cropped ones)")

	flag.BoolVar(&flags.TLS, "tls", false, "should it listen on TLS (443)")
	flag.StringVar(&flags.TLSDomain, "domain", "", "tls domain")
	flag.IntVar(&flags.ListenPort, "port", 8080, "port to listen on (if not tls)")

	flag.Parse()

	return flags
}