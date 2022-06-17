package main

import (
	"fmt"
	"image"
	"os"
	"time"

	"github.com/corona10/goimagehash"
	"github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"

	config "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/config/automatorconfig"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/imgutils"
	schema "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/ocrschema"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/rokocr"
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
	go func() {
		for {
			time.Sleep(time.Millisecond * 500)
			img, err := screencapture(device)
			if err != nil {
				log.Errorf("Error: %v", err)
				continue
			}

			processImage(img)
		}
	}()

	<-quit

	return nil
}

func processImage(img image.Image) {
	imagehash, _ := goimagehash.DifferenceHash(img)
	log.Infof("handling image: %vx%v, hash: %x", img.Bounds().Dx(), img.Bounds().Dy(), imagehash.GetHash())

	// re-read templates white testing
	powerDetailsTemplate, _ := schema.LoadTemplate("./templates/iphone-11-with-power.json")
	powerRatingsTemplate, _ := schema.LoadTemplate("./automation/power-ratings.json")
	profileTemplate, _ := schema.LoadTemplate("./automation/profile.json")
	rankingsSelection, _ := schema.LoadTemplate("./automation/rankings-selection.json")

	// detect screen based on template
	if powerDetailsTemplate.Matches(img) {
		result := rokocr.ParseImage("power_details", img, powerDetailsTemplate, flags.TmpDirectory, flags.TessdataDirectory)
		log.Infof("Detected power details screen: %v", pp.Sprint(result))
	} else if profileTemplate.Matches(img) {
		log.Debugf("Detected profile screen")
	} else if rankingsSelection.Matches(img) {
		log.Infof("Detected rankings screen")
	} else if powerRatingsTemplate.Matches(img) {
		log.Debugf("Detected power ratings screen")
	} else {
		log.Debugf("Unknown screen: %x", imagehash.GetHash())
		imgutils.WritePNGImage(img, fmt.Sprintf("./media/unknown_%x.png", imagehash.GetHash()))
	}
}
