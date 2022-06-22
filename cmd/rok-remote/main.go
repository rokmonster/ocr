package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"net/url"
	"os"
	"sync"

	"github.com/corona10/goimagehash"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

		rc := RemoteControllerClient{
			device: device,
			server: flags.ROKServer,
		}

		go func() {
			defer wg.Done()
			rc.DeviceRegisterAndWork()
		}()
	}

	if len(serials) == 0 {
		log.Fatalf("No devices found. Check your connection, or try `adb devices`")
	}

	// wait for all devices to finish
	wg.Wait()
}

type RemoteControllerClient struct {
	device *adb.Device
	server string
}

func (c *RemoteControllerClient) getWebsocketURI() url.URL {
	uri, _ := url.Parse(c.server)
	scheme := uri.Scheme

	if uri.Scheme == "https" {
		scheme = "wss"
	}

	if uri.Scheme == "http" {
		scheme = "ws"
	}

	return url.URL{Scheme: scheme, Host: uri.Host, Path: "/devices/ws"}
}

func (c *RemoteControllerClient) DeviceRegisterAndWork() {
	info, _ := c.device.DeviceInfo()
	defer log.Warnf("Done with device: %v", info.Serial)
	log.Infof("Device: %v", info.Serial)

	u := c.getWebsocketURI()
	log.Infof("Connecting to %s", u.String())

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	defer func() {
		log.Debugf("Sending bye bye...")
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))
		ws.Close()
	}()

	log.Infof("Saying hello: %v", info.Serial)

	_ = ws.WriteJSON(gin.H{
		"serial": info.Serial,
	})

	// read instruction & handle it
handlerloop:
	for {
		var message struct {
			Command string      `json:"command"`
			Args    interface{} `json:"args,omitempty"`
		}

		err := ws.ReadJSON(&message)
		if err != nil {
			log.Errorf("read: %v", err)
			break handlerloop
		}

		switch message.Command {
		case "quit":
			break handlerloop
		case "imagehash":
			c.doImageHash(ws)
		case "image":
			c.doSendImage(ws)
		default:
			log.Warnf("Unknown command: '%v' received", message.Command)
		}
	}

	log.Infof("We broke from handler loop...")
}

func (c *RemoteControllerClient) doImageHash(ws *websocket.Conn) {
	img, _ := screencapture(c.device)
	imagehash, _ := goimagehash.DifferenceHash(img)
	log.Infof("[imagehash] w: %v, h: %v, hash: %x", img.Bounds().Dx(), img.Bounds().Dy(), imagehash.GetHash())
	ws.WriteJSON(gin.H{
		"responseType": "imagehash",
		"value": gin.H{
			"hash": fmt.Sprintf("%x", imagehash.GetHash()),
			"w":    img.Bounds().Dx(),
			"h":    img.Bounds().Dy(),
		},
	})
}

func (c *RemoteControllerClient) doSendImage(ws *websocket.Conn) {
	img, _ := screencapture(c.device)
	buf := new(bytes.Buffer)

	png.Encode(buf, img)

	log.Infof("[image] w: %v, h: %v, len: %v bytes", img.Bounds().Dx(), img.Bounds().Dy(), buf.Len())
	ws.WriteJSON(gin.H{
		"responseType": "image",
		"value":        buf.Bytes(),
	})
}

func screencapture(device *adb.Device) (image.Image, error) {
	// screencap
	cmdOutput, err := device.RunCommand("screencap -p")
	if err != nil {
		return nil, err
	}
	_ = os.WriteFile("out.png", []byte(cmdOutput), 0644)

	// read image
	return imgutils.ReadImage("out.png")
}
