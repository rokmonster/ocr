package fileutils

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func Mkdirs(path string) {
	p, _ := filepath.Abs(path)
	log.Infof("Creating dir: %s", p)
	os.MkdirAll(p, os.ModePerm)
}
