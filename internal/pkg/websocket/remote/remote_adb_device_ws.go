package remote

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"

	"github.com/corona10/goimagehash"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rokmonster/ocr/internal/pkg/imgutils"
	log "github.com/sirupsen/logrus"
	adb "github.com/zach-klippenstein/goadb"
)

type RemoteADBDeviceWS struct {
	device    *adb.Device
	lastImage *image.Image
	wsUri     string
}

func NewRemoteADBDeviceWS(websocketUri string, device *adb.Device) *RemoteADBDeviceWS {
	return &RemoteADBDeviceWS{
		wsUri:  websocketUri,
		device: device,
	}
}

func (c *RemoteADBDeviceWS) DeviceRegisterAndWork() {
	info, _ := c.device.DeviceInfo()
	defer log.Warnf("Done with device: %v", info.Serial)

	log.Infof("Device: %v", info.Serial)
	log.Infof("Connecting to %s", c.wsUri)

	ws, _, err := websocket.DefaultDialer.Dial(c.wsUri, nil)
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
			Command string          `json:"command"`
			Args    json.RawMessage `json:"args"`
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
		case "tap":
			var value struct {
				X int `json:"x"`
				Y int `json:"y"`
			}
			json.Unmarshal(message.Args, &value)
			log.Infof("[tap] x: %v, y: %v", value.X, value.Y)
			c.doTap(ws, value.X, value.Y)
		default:
			log.Warnf("[%v] received with args: %+v", message.Command, message.Args)
		}
	}

	log.Infof("We broke from handler loop...")
}

func (c *RemoteADBDeviceWS) doImageHash(ws *websocket.Conn) {
	img, _ := c.screencapture()
	imagehash, _ := goimagehash.DifferenceHash(img)
	c.lastImage = &img
	log.Infof("[imagehash (newimage)] w: %v, h: %v, hash: %x", img.Bounds().Dx(), img.Bounds().Dy(), imagehash.GetHash())
	ws.WriteJSON(gin.H{
		"responseType": "imagehash",
		"value": gin.H{
			"hash": imagehash.GetHash(),
			"w":    img.Bounds().Dx(),
			"h":    img.Bounds().Dy(),
		},
	})
}

func (c *RemoteADBDeviceWS) doTap(ws *websocket.Conn, x, y int) {
	_, e := c.device.RunCommand("input", "tap", fmt.Sprintf("%v", x), fmt.Sprintf("%v", y))
	if e != nil {
		log.Errorf("[tap] failed with: %v", e)
	}
}

func (c *RemoteADBDeviceWS) doSendImage(ws *websocket.Conn) error {
	if c.lastImage == nil {
		return errors.New("can't send the image, no image captured yet")
	}
	img := *c.lastImage

	buf := new(bytes.Buffer)
	png.Encode(buf, img)

	log.Infof("[image] w: %v, h: %v, len: %v bytes", img.Bounds().Dx(), img.Bounds().Dy(), buf.Len())
	return ws.WriteJSON(gin.H{
		"responseType": "image",
		"value":        buf.Bytes(),
	})
}

func (c *RemoteADBDeviceWS) screencapture() (image.Image, error) {
	// screencap
	cmdOutput, err := c.device.RunCommand("screencap -p")
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBufferString(cmdOutput)
	return imgutils.ReadImage(buf)
}
