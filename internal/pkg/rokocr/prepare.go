package rokocr

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rokmonster/ocr/templates"
	"github.com/sirupsen/logrus"

	"github.com/rokmonster/ocr/internal/pkg/utils/fileutils"

	"github.com/rokmonster/ocr/internal/pkg/config"
	"github.com/rokmonster/ocr/internal/pkg/config/serverconfig"
)

func InstallSystemD(flags serverconfig.ROKServerConfig) {
	workingDir := fmt.Sprintf("/home/%v", flags.InstallUser)
	if flags.InstallUser == "root" {
		workingDir = "/root"
	}

	fileutils.WriteFile([]byte(fmt.Sprintf("%s=%v\n%s=%v\n",
		"OAUTH_CLIENT_ID", flags.OAuthClientID,
		"OAUTH_SECRET_ID", flags.OAuthSecretID,
	)), "/etc/rokmonster.env")
	os.Chown("/etc/rokmonster.env", 0, 0)
	os.Chmod("/etc/rokmonster.env", 0600)

	fileutils.WriteFile([]byte(fmt.Sprintf(`[Unit]
Description=ROKMonster OCR Server
Requires=rokocr-server-https.socket
Requires=rokocr-server-http.socket
After=syslog.target
After=network.target

[Service]
RestartSec=2s
Type=simple
User=%v
Group=%v
WorkingDirectory=%v
EnvironmentFile=/etc/rokmonster.env
ExecStart=/usr/bin/rok-server -tls -domain %v
Restart=always

[Install]
WantedBy=multi-user.target`, flags.InstallUser, flags.InstallUser, workingDir, flags.TLSDomain)), "/etc/systemd/system/rokocr-server.service")

	fileutils.WriteFile([]byte(`[Socket]
ListenStream=443
NoDelay=true
FileDescriptorName=https
Service=rokocr-server.service

[Install]
WantedBy = sockets.target`), "/etc/systemd/system/rokocr-server-https.socket")

	fileutils.WriteFile([]byte(`[Socket]
ListenStream=80
NoDelay=true
FileDescriptorName=http
Service=rokocr-server.service

[Install]
WantedBy = sockets.target`), "/etc/systemd/system/rokocr-server-http.socket")

}

func Prepare(flags config.CommonConfiguration) {
	fileutils.Mkdirs(flags.TessdataDirectory)
	fileutils.Mkdirs(flags.MediaDirectory)
	fileutils.Mkdirs(flags.TemplatesDirectory)

}

func PreloadTemplates(flags config.CommonConfiguration) {
	files, _ := templates.FS.ReadDir(".")
	base, _ := filepath.Abs(flags.TemplatesDirectory)

	for _, r := range files {
		path := filepath.Join(base, r.Name())
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			logrus.Infof("Preloading template: %v => %v", r.Name(), path)
			b, _ := templates.FS.ReadFile(r.Name())
			_ = fileutils.WriteFile(b, path)
		}
	}
}

func DownloadTesseractData(flags config.CommonConfiguration) {
	langFiles := []string{
		"eng",     // English
		"rus",     // Russian
		"fra",     // French
		"spa",     // Spanish
		"chi_tra", // Chinese Traditional
		"chi_sim", // Chinese Simplified
		"jpn",     // Japan
		"ita",     // Italian
		"kor",     // Korean
	}

	for _, lang := range langFiles {
		path := filepath.Join(flags.TessdataDirectory, fmt.Sprintf("%v.traineddata", lang))
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			_ = fileutils.Download(path, fmt.Sprintf("https://raw.githubusercontent.com/tesseract-ocr/tessdata/main/%v.traineddata", lang))
		}
	}
}
