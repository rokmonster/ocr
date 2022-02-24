package fileutils

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func Mkdirs(path string) {
	p, _ := filepath.Abs(path)
	logrus.Infof("Creating dir: %s", p)
	os.MkdirAll(p, os.ModePerm)
}
