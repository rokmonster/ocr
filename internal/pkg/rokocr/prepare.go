package rokocr

import (
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/config"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/config/serverconfig"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/fileutils"
)

func InstallSystemD(flags serverconfig.RokServerConfiguration) {
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
WorkingDirectory=/home/%v
ExecStart=/usr/bin/rok-server -tls -domain %v
Restart=always

[Install]
WantedBy=multi-user.target`, flags.InstallUser, flags.InstallUser, flags.InstallUser, flags.TLSDomain)), "/etc/systemd/system/rokocr-server.service")

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

	if len(fileutils.GetFilesInDirectory(flags.TessdataDirectory)) == 0 {
		logrus.Warnf("No tesseract trained data found, downloading english & french ones")
		fileutils.Download(filepath.Join(flags.TessdataDirectory, "eng.traineddata"), "https://github.com/tesseract-ocr/tessdata/raw/main/eng.traineddata")
		fileutils.Download(filepath.Join(flags.TessdataDirectory, "fra.traineddata"), "https://github.com/tesseract-ocr/tessdata/raw/main/fra.traineddata")
		fileutils.Download(filepath.Join(flags.TessdataDirectory, "spa.traineddata"), "https://github.com/tesseract-ocr/tessdata/raw/main/spa.traineddata")
	}
}
