package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/corona10/goimagehash"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/olekukonko/tablewriter"
	"github.com/rokmonster/ocr/internal/pkg/ocrschema"
	"github.com/rokmonster/ocr/internal/pkg/rokocr"
	"github.com/rokmonster/ocr/internal/pkg/stringutils"
	log "github.com/sirupsen/logrus"
)

func NewRemoteServerWS(device *RemoteServerClient, socket *websocket.Conn, templates, tessdata string) *remoteServerWS {
	return &remoteServerWS{
		results:      make([]ocrschema.OCRResponse, 0),
		templatesDir: templates,
		tessdataDir:  tessdata,
		device:       device,
		socket:       socket,
	}
}

type remoteServerWS struct {
	device       *RemoteServerClient
	socket       *websocket.Conn
	templatesDir string
	tessdataDir  string
	results      []ocrschema.OCRResponse
}

func (c *remoteServerWS) requestScreenHash() {
	c.socket.WriteJSON(gin.H{"command": "imagehash"})
}

func (c *remoteServerWS) requestImage() {
	c.socket.WriteJSON(gin.H{"command": "image"})
}

func (c *remoteServerWS) requestDisconnect() {
	c.socket.WriteJSON(gin.H{"command": "quit"})
}

func (c *remoteServerWS) requestTap(x, y int) {
	c.socket.WriteJSON(gin.H{"command": "tap", "args": gin.H{
		"x": x,
		"y": y,
	}})

	// wait a few seconds
	time.Sleep(time.Second * 2)
}

func (c *remoteServerWS) processImage(img image.Image) {
	templates := rokocr.LoadTemplates(c.templatesDir)
	imagehash, _ := goimagehash.DifferenceHash(img)
	t := rokocr.PickTemplate(imagehash, templates)
	// TODO: Check if match???
	res := rokocr.ParseImage(fmt.Sprintf("remoteimage_%v.png", time.Now().Format("20060102_150405")), img, t, os.TempDir(), c.tessdataDir)

	// put results
	c.results = append(c.results, res)
	c.PrintAll()

	// todo where to TAP?
	c.requestTap(100, 300)
}

func (c *remoteServerWS) PrintAll() {
	headers := []string{}
	for _, r := range c.results {
		for k := range r.Data {
			headers = append(headers, k)
		}
	}

	headers = stringutils.Unique(headers)

	output := new(bytes.Buffer)

	table := tablewriter.NewWriter(output)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeader(headers)
	for _, row := range c.results {
		rowData := []string{}

		for _, x := range headers {
			if value, ok := row.Data[x]; ok {
				rowData = append(rowData, fmt.Sprintf("%v", value))
			} else {
				rowData = append(rowData, "")
			}
		}

		table.Append(rowData)
	}

	table.Render()
	log.Infof("[%v] Results so far: \n%v", c.device.Name, output.String())
}

func (c *remoteServerWS) isScreenInteresting(w, h int, hash *goimagehash.ImageHash) bool {
	templates := rokocr.LoadTemplates(c.templatesDir)
	t := rokocr.PickTemplate(hash, templates)
	return t.Match(hash)
}

func (c *remoteServerWS) Loop() {
	c.requestScreenHash()

	// read message, send command, etc...
	for {
		var wsResponse struct {
			ResponseType string          `json:"responseType"`
			Value        json.RawMessage `json:"value"`
		}
		err := c.socket.ReadJSON(&wsResponse)
		if err != nil {
			log.Error(err)
			break
		}
		switch wsResponse.ResponseType {
		case "imagehash":
			{
				var value struct {
					Height int    `json:"h"`
					Width  int    `json:"w"`
					Hash   uint64 `json:"hash"`
				}
				json.Unmarshal(wsResponse.Value, &value)
				hash := goimagehash.NewImageHash(value.Hash, goimagehash.DHash)

				log.Infof("[%v] on screen with hash: %x", c.device.Name, hash.GetHash())

				if c.isScreenInteresting(value.Width, value.Height, hash) {
					c.requestImage()
				} else {
					// ask for new screenshot after 1? seconds
					time.Sleep(time.Second * 1)
					c.requestScreenHash()
				}
			}

		case "image":
			{
				var msgBytes []byte
				json.Unmarshal(wsResponse.Value, &msgBytes)
				img, _ := png.Decode(bytes.NewReader(msgBytes))
				log.Infof("[%v] sent image of size: %v bytes => w: %v h: %v", c.device.Name, len(msgBytes), img.Bounds().Dx(), img.Bounds().Dy())

				c.processImage(img)
				c.requestScreenHash()
			}
		case "error":
			{
				c.requestDisconnect()
			}
		default:
			{
				log.Errorf("[%v] unknown response received: %+v", c.device.Name, wsResponse.ResponseType)
			}
		}
	}
}
