package fileutils

import (
	"io"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

func Download(filepath string, url string) error {
	logrus.Infof("Downloading: %v", url)
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
