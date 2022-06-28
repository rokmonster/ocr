package imgutils

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
)

func ReadImageFile(filename string) (image.Image, error) {
	imgfile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer imgfile.Close()

	return ReadImage(imgfile)
}

func ReadImage(reader io.Reader) (image.Image, error) {
	// try to decode as PNG
	img, _, err := image.Decode(reader)
	if err == nil {
		return img, nil
	}

	return nil, fmt.Errorf("unsupported file format")
}
