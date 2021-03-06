package serverconfig

import (
	"flag"
	"os"

	"github.com/rokmonster/ocr/internal/pkg/config"
)

type ROKServerConfig struct {
	config.CommonConfiguration
	ListenPort    int
	TLS           bool
	TLSDomain     string
	Install       bool
	InstallUser   string
	OAuthClientID string
	OAuthSecretID string
}

func Parse() ROKServerConfig {
	var flags ROKServerConfig

	flag.StringVar(&flags.MediaDirectory, "media", "./media", "folder where all files to scan is placed")
	flag.StringVar(&flags.TemplatesDirectory, "templates", "./templates", "templates dir")
	flag.StringVar(&flags.TessdataDirectory, "tessdata", "./tessdata", "tesseract data files directory")
	flag.StringVar(&flags.OutputDirectory, "output", "./out", "output dir")
	flag.StringVar(&flags.TmpDirectory, "tmp", os.TempDir(), "Directory for temporary files (cropped ones)")
	flag.BoolVar(&flags.Install, "install", false, "Create systemd unit and exits")
	flag.StringVar(&flags.InstallUser, "user", "root", "which user to install")
	flag.BoolVar(&flags.TLS, "tls", false, "should it listen on TLS (443)")
	flag.StringVar(&flags.TLSDomain, "domain", "", "tls domain")
	flag.IntVar(&flags.ListenPort, "port", 8080, "port to listen on (if not tls)")

	flag.StringVar(&flags.OAuthClientID, "oauth-clientid", os.Getenv("OAUTH_CLIENT_ID"), "Google OAuth Client ID")
	flag.StringVar(&flags.OAuthSecretID, "oauth-secretid", os.Getenv("OAUTH_SECRET_ID"), "Google OAuth Secret ID")

	flag.Parse()

	return flags
}
