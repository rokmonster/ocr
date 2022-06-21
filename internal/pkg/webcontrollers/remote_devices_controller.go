package webcontrollers

import (
	"bytes"
	"encoding/json"
	"image/png"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type RemoteDevice struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

func (controller *RemoteDevicesController) getRemoteDevices() []RemoteDevice {
	devices := []RemoteDevice{}

	for _, d := range controller.clients {
		devices = append(devices, d)
	}

	return devices
}

type RemoteDevicesController struct {
	router   *gin.RouterGroup
	clients  map[*websocket.Conn]RemoteDevice
	upgrader websocket.Upgrader
}

func NewRemoteDevicesController(router *gin.RouterGroup) *RemoteDevicesController {
	return &RemoteDevicesController{
		router:  router,
		clients: make(map[*websocket.Conn]RemoteDevice),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (controller *RemoteDevicesController) Setup() {
	controller.router.GET("/", func(c *gin.Context) {
		data := gin.H{
			"devices": controller.getRemoteDevices(),
		}

		switch c.NegotiateFormat(gin.MIMEJSON, gin.MIMEHTML) {
		case gin.MIMEHTML:
			c.HTML(http.StatusOK, "devices.html", data)
		case gin.MIMEJSON:
			c.JSON(http.StatusOK, data)
		}
	})

	controller.router.GET("/ws", func(c *gin.Context) {
		ws, _ := controller.upgrader.Upgrade(c.Writer, c.Request, nil)

		// don't forget to close the connection & remove client
		defer ws.Close()
		defer delete(controller.clients, ws)

		// first message on WS should be our hello
		var deviceInfo struct {
			Serial string `json:"serial"`
		}
		err := ws.ReadJSON(&deviceInfo)

		if err != nil {
			logrus.Errorf("I don't like this WS Client: %v", err)
			ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "I expect you to behave nicely"))
			return
		} else {
			logrus.Infof("Device connected from: %v => %v", ws.RemoteAddr(), deviceInfo.Serial)

			device := RemoteDevice{
				Address: ws.RemoteAddr().String(),
				Name:    deviceInfo.Serial,
			}
			// register handler && start the loop
			handler := WSRemoteHandler{socket: ws, device: &device}

			// put the client into active clients...
			controller.clients[ws] = device

			// handle the command / send actions
			handler.Loop()
		}
	})
}

type WSRemoteHandler struct {
	device *RemoteDevice
	socket *websocket.Conn
}

func (c *WSRemoteHandler) requestScreenHash() {
	c.socket.WriteJSON(gin.H{"command": "imagehash"})
}

func (c *WSRemoteHandler) requestImage() {
	c.socket.WriteJSON(gin.H{"command": "image"})
}

func (c *WSRemoteHandler) requestDisconnect() {
	c.socket.WriteJSON(gin.H{"command": "quit"})
}

func (c *WSRemoteHandler) Loop() {
	c.requestScreenHash()

	// read message, send command, etc...
	for {
		var wsResponse struct {
			ResponseType string          `json:"responseType"`
			Value        json.RawMessage `json:"value"`
		}
		err := c.socket.ReadJSON(&wsResponse)
		if err != nil {
			logrus.Error(err)
			break
		}
		logrus.Infof("Received message from: %v => type: %+v", c.device.Address, wsResponse.ResponseType)
		switch wsResponse.ResponseType {
		case "imagehash":
			{
				var value struct {
					Height int    `json:"h"`
					Width  int    `json:"w"`
					Hash   string `json:"hash"`
				}
				json.Unmarshal(wsResponse.Value, &value)
				logrus.Infof("We are on screen with hash: %+v", value)
				// ask for new screenshot after 2 seconds
				time.Sleep(time.Second * 2)
				c.requestScreenHash()
			}

		case "image":
			{
				// png.Decode(bufio.NewReader(wsResponse.Value))
				var msgBytes []byte
				json.Unmarshal(wsResponse.Value, &msgBytes)
				img, _ := png.Decode(bytes.NewReader(msgBytes))
				logrus.Infof("we got full image of size: %v bytes => w: %v h: %v", len(msgBytes), img.Bounds().Dx(), img.Bounds().Dy())
				c.requestDisconnect()
			}
		}
	}
}
