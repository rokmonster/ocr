package imgutils

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
)

func ReadImage(filename string) (image.Image, error) {
	imgfile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer imgfile.Close()

	// try to decode as PNG
	img, err := png.Decode(imgfile)
	if err == nil {
		return img, nil
	}

	// try to decode as PNG
	img, err = jpeg.Decode(imgfile)
	if err == nil {
		return img, nil
	}

	return nil, fmt.Errorf("unsupported file format")
}
