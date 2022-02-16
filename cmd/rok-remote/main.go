package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/corona10/goimagehash"
	log "github.com/sirupsen/logrus"

	"github.com/k0kubun/pp"
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

	waitExitProfile := false

	ticker := time.NewTicker(1 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				// re-read templates white testing
				powerDetailsTemplate := schema.LoadTemplate("./templates/iphone-11-with-power.json")
				powerRatingsTemplate := schema.LoadTemplate("./templates/iphone-11-power-rankings.json")
				profileTemplate := schema.LoadTemplate("./templates/iphone-11-profile.json")

				// screencap
				cmdOutput, err := device.RunCommand("screencap -p")
				if err != nil {
					fmt.Println("\terror running command:", err)
				}
				_ = os.WriteFile("out.png", []byte(cmdOutput), 0644)

				// read image
				img, err := imgutils.ReadImage("out.png")
				if err != nil {
					continue
				}

				// detect screen based on template
				if powerDetailsTemplate.Matches(img) {
					log.Infof("Detected power details screen - parsing")
					result := rokocr.ParseImage("out.png", img, &powerDetailsTemplate, os.TempDir(), "./tessdata")
					pp.Println(result)
					waitExitProfile = false
				} else if profileTemplate.Matches(img) {
					if !waitExitProfile {
						log.Infof("Detected profile screen")
						device.RunCommand("input touchscreen tap 1850 125")
						waitExitProfile = true
					}
				} else if powerRatingsTemplate.Matches(img) {
					log.Infof("Detected power ratings screen")
					result := rokocr.ParseImage("out.png", img, &powerRatingsTemplate, os.TempDir(), "./tessdata")
					pp.Println(result)
					device.RunCommand("input touchscreen tap 1140 980")
					waitExitProfile = false
				} else {
					imagehash, _ := goimagehash.DifferenceHash(img)
					log.Warnf("Unknown screen: %x", imagehash.GetHash())
					waitExitProfile = false
				}

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	<-quit

	return nil
}
