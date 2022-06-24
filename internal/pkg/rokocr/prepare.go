package rokocr

import (
	"errors"
	"fmt"
	"github.com/rokmonster/ocr/internal/pkg/utils/fileutils"
	"os"
	"path/filepath"

	"github.com/rokmonster/ocr/internal/pkg/config"
	"github.com/rokmonster/ocr/internal/pkg/config/serverconfig"
)

func InstallSystemD(flags serverconfig.ROKServerConfig) {
	workingDir := fmt.Sprintf("/home/%v", flags.InstallUser)
	if flags.InstallUser == "root" {
		workingDir = "/root"
	}

	fileutils.WriteFile([]byte(fmt.Sprintf(`[Unit]
Description=ROK OCR Server
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
			_ = fileutils.Download(path, fmt.Sprintf("https://github.com/tesseract-ocr/tessdata/raw/main/%v.traineddata", lang))
		}
	}
}
