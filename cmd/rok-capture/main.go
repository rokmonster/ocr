package main

import (
	"fmt"
	"image"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	config "github.com/rokmonster/ocr/internal/pkg/config/automatorconfig"
	"github.com/rokmonster/ocr/internal/pkg/imgutils"
	"github.com/rokmonster/ocr/internal/pkg/rokocr"
	adb "github.com/zach-klippenstein/goadb"
)

var flags = config.Parse()

var (
	client *adb.Adb
)

func main() {
	rokocr.Prepare(flags.CommonConfiguration)

	var err error
	client, err = adb.NewWithConfig(adb.ServerConfig{
		Port: flags.ADBPort,
	})

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Starting adb server")
	client.StartServer()

	serverVersion, err := client.ServerVersion()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server version:", serverVersion)

	ConnectAndUseDevice(adb.AnyDevice())
}

func ConnectAndUseDevice(descriptor adb.DeviceDescriptor) {
	device := client.Device(descriptor)
	if err := workWithDevice(device); err != nil {
		log.Println(err)
	}
}

func screencapture(device *adb.Device) (image.Image, error) {
	// screencap
	cmdOutput, err := device.RunCommand("screencap -p")
	if err != nil {
		fmt.Println("\terror running command:", err)
	}
	_ = os.WriteFile("out.png", []byte(cmdOutput), 0644)

	// read image
	return imgutils.ReadImage("out.png")
}

func workWithDevice(device *adb.Device) error {
	serialNo, err := device.Serial()
	if err != nil {
		return err
	}
	devPath, err := device.DevicePath()
	if err != nil {
		return err
	}

	log.Printf("serial no: %s", serialNo)
	log.Printf("devPath: %s", devPath)

	quit := make(chan struct{})
	i := 0
	go func() {
		screens := []string{"profile", "profile_with_stats", "more_details"}
		for {
			for _, screen := range screens {
				fileName, _ := filepath.Abs(fmt.Sprintf("%v/%04d_%v.png", flags.MediaDirectory, i, screen))
				log.Infof("Press ENTER to capture: %v", fileName)
				fmt.Scanln()
				img, err := screencapture(device)
				if err != nil {
					log.Errorf("Error: %v", err)
					continue
				}
				imgutils.WritePNGImage(img, fileName)
			}
			i++
		}
	}()

	<-quit

	return nil
}
