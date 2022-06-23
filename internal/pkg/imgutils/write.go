package imgutils

import (
	"image"
	"image/png"
	"os"

	log "github.com/sirupsen/logrus"
)

// WritePNGImage - writes an Image back to the disk.
func WritePNGImage(img image.Image, name string) error {
	fd, err := os.Create(name)
	if err != nil {
		log.Errorf("failed to write: %v", err)
		return err
	}
	defer fd.Close()

	return png.Encode(fd, img)
}
