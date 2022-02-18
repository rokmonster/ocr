package main

import (
	"flag"
	"fmt"
	"image"
	"os"
	"time"

	"github.com/corona10/goimagehash"
	log "github.com/sirupsen/logrus"

	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/imgutils"
	schema "github.com/xor22h/rok-monster-ocr-golang/internal/pkg/ocrschema"
	"github.com/xor22h/rok-monster-ocr-golang/internal/pkg/rokocr"
	adb "github.com/zach-klippenstein/goadb"
)

var (
	port = flag.Int("p", adb.AdbPort, "")

	client *adb.Adb
)

func main() {
	flag.Parse()

	var err error
	client, err = adb.NewWithConfig(adb.ServerConfig{
		Port: *port,
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
	if err := Top300Loop(device); err != nil {
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

func Top300Loop(device *adb.Device) error {
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

	// re-read templates white testing
	powerDetailsTemplate := schema.LoadTemplate("./templates/iphone-11-with-power.json")
	powerRatingsTemplate := schema.LoadTemplate("./automation/power-ratings.json")
	profileTemplate := schema.LoadTemplate("./automation/profile.json")
	rankingsSelection := schema.LoadTemplate("./automation/rankings-selection.json")

	img, err := screencapture(device)
	quit := make(chan struct{})
	go func() {
		for {
			if err != nil {
				log.Errorf("Error: %v", err)
				time.Sleep(time.Second * 1)
				img, err = screencapture(device)
				continue
			}

			// detect screen based on template
			if powerDetailsTemplate.Matches(img) {
				log.Infof("Detected power details screen")
				img, err = screencapture(device)

			} else if profileTemplate.Matches(img) {
				log.Debugf("Detected profile screen")
				// close the screen
				device.RunCommand("input touchscreen tap 1850 125")
				time.Sleep(time.Millisecond * 500)
				img, err = screencapture(device)
			} else if rankingsSelection.Matches(img) {
				log.Infof("Detected rankings screen")
				time.Sleep(time.Millisecond * 500)
				img, err = screencapture(device)
			} else if powerRatingsTemplate.Matches(img) {
				log.Debugf("Detected power ratings screen")
				result := rokocr.ParseImage("out.png", img, &powerRatingsTemplate, os.TempDir(), "./tessdata")
				log.Printf("Power ratings: %v -> %v", result["place_4"], result["place_4_value"])
				log.Printf("Power ratings: %v -> %v", result["place_5"], result["place_5_value"])
				log.Printf("Power ratings: %v -> %v", result["place_6"], result["place_6_value"])
				device.RunCommand("input touchscreen tap 1140 980")
				time.Sleep(time.Millisecond * 500)
				img, err = screencapture(device)
			} else {
				imagehash, _ := goimagehash.DifferenceHash(img)
				log.Debugf("Unknown screen: %x", imagehash.GetHash())
				imgutils.WritePNGImage(img, fmt.Sprintf("./media/unknown_%x.png", imagehash.GetHash()))
				time.Sleep(time.Millisecond * 500)
				img, err = screencapture(device)
			}
		}
	}()

	<-quit

	return nil
}
