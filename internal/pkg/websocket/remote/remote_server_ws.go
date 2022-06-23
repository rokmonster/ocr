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

func NewRemoteServerWS(socket *websocket.Conn, device, templates, tessdata string) *RemoteServerWS {
	return &RemoteServerWS{
		results:      make([]ocrschema.OCRResponse, 0),
		templatesDir: templates,
		tessdataDir:  tessdata,
		device:       device,
		socket:       socket,
	}
}

type RemoteServerWS struct {
	socket       *websocket.Conn
	device       string
	templatesDir string
	tessdataDir  string
	results      []ocrschema.OCRResponse
}

func (c *RemoteServerWS) Disconnect() error {
	return c.requestDisconnect()
}

func (c *RemoteServerWS) GetData() []ocrschema.OCRResponse {
	return c.results
}

func (c *RemoteServerWS) requestScreenHash() error {
	return c.socket.WriteJSON(gin.H{"command": "imagehash"})
}

func (c *RemoteServerWS) requestImage() error {
	return c.socket.WriteJSON(gin.H{"command": "image"})
}

func (c *RemoteServerWS) requestDisconnect() error {
	return c.socket.WriteJSON(gin.H{"command": "quit"})
}

func (c *RemoteServerWS) requestTap(x, y int) error {
	return c.socket.WriteJSON(gin.H{"command": "tap", "args": gin.H{
		"x": x,
		"y": y,
	}})
}

func (c *RemoteServerWS) processImage(img image.Image) {
	templates := rokocr.LoadTemplates(c.templatesDir)
	imagehash, _ := goimagehash.DifferenceHash(img)
	t := rokocr.PickTemplate(imagehash, templates)
	// TODO: Check if match???
	res := rokocr.ParseImage(fmt.Sprintf("remoteimage_%v.png", time.Now().Format("20060102_150405")), img, t, os.TempDir(), c.tessdataDir)

	// put results
	c.results = append(c.results, res)
	c.printAll()

	// todo where to TAP?
	_ = c.requestTap(100, 300)
	// wait a few seconds
	time.Sleep(time.Second * 2)
}

func (c *RemoteServerWS) printAll() {
	var headers []string
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
		var rowData []string

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
	log.Infof("[%v] Results so far: \n%v", c.device, output.String())
}

func (c *RemoteServerWS) isScreenInteresting(w, h int, hash *goimagehash.ImageHash) bool {
	templates := rokocr.LoadTemplates(c.templatesDir)
	t := rokocr.PickTemplate(hash, templates)
	return t.Match(hash)
}

func (c *RemoteServerWS) Loop() {
	_ = c.requestScreenHash()

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

				log.Infof("[%v] on screen with hash: %x", c.device, hash.GetHash())

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
				log.Infof("[%v] sent image of size: %v bytes => w: %v h: %v", c.device, len(msgBytes), img.Bounds().Dx(), img.Bounds().Dy())

				c.processImage(img)
				c.requestScreenHash()
			}
		case "error":
			{
				c.requestDisconnect()
			}
		default:
			{
				log.Errorf("[%v] unknown response received: %+v", c.device, wsResponse.ResponseType)
			}
		}
	}
}
