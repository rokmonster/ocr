package main

import (
	"net/url"
	"sync"

	"github.com/rokmonster/ocr/internal/pkg/utils"

	config "github.com/rokmonster/ocr/internal/pkg/config/rokremoteconfig"
	wsclient "github.com/rokmonster/ocr/internal/pkg/websocket/client"
	log "github.com/sirupsen/logrus"
	adb "github.com/zach-klippenstein/goadb"
)

var flags = config.Parse()

var (
	client *adb.Adb
)

func main() {
	var err error
	client, err = adb.NewWithConfig(adb.ServerConfig{
		Port: flags.ADBPort,
	})
	utils.Panic(err)

	log.Println("Starting adb server")
	utils.Panic(client.StartServer())

	serverVersion, err := client.ServerVersion()
	utils.Panic(err)

	log.Println("ADB Server version:", serverVersion)

	// register all available devices with remote
	var wg sync.WaitGroup

	serials, _ := client.ListDeviceSerials()
	for _, serial := range serials {
		wg.Add(1)
		descriptor := adb.DeviceWithSerial(serial)
		device := client.Device(descriptor)
		info, _ := device.DeviceInfo()

		log.Infof("Found device: %v - %v", serial, info.DeviceInfo)

		uri := getWebsocketURI(flags.Server)
		rc := wsclient.NewADBDeviceWS(uri.String(), device)

		go func() {
			defer wg.Done()
			rc.DeviceRegisterAndWork() // this one is blocking...
		}()
	}

	if len(serials) == 0 {
		log.Fatalf("No devices found. Check your connection, or try `adb devices`")
	}

	// wait for all devices to finish
	wg.Wait()
}

func getWebsocketURI(server string) url.URL {
	uri, _ := url.Parse(server)
	scheme := uri.Scheme

	if uri.Scheme == "https" {
		scheme = "wss"
	}

	if uri.Scheme == "http" {
		scheme = "ws"
	}

	return url.URL{Scheme: scheme, Host: uri.Host, Path: "/devices/ws"}
}
