package fileutils

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func WriteFile(data []byte, name string) error {
	fd, err := os.Create(name)
	if err != nil {
		log.Errorf("failed to write: %v", err)
		return err
	}
	defer fd.Close()

	_, err = fd.Write(data)
	return err
}
