package webcontrollers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rokmonster/ocr/internal/pkg/websocket/remote"
	log "github.com/sirupsen/logrus"
)

func (controller *RemoteDevicesController) getRemoteDevices() []remote.RemoteServerClient {
	devices := []remote.RemoteServerClient{}

	for _, d := range controller.clients {
		devices = append(devices, d)
	}

	return devices
}

type RemoteDevicesController struct {
	router       *gin.RouterGroup
	clients      map[*websocket.Conn]remote.RemoteServerClient
	upgrader     websocket.Upgrader
	templatesDir string
	tessdataDir  string
}

func NewRemoteDevicesController(router *gin.RouterGroup, templates, tessdata string) *RemoteDevicesController {
	return &RemoteDevicesController{
		router:       router,
		clients:      make(map[*websocket.Conn]remote.RemoteServerClient),
		templatesDir: templates,
		tessdataDir:  tessdata,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// who care's about CORS?
				// P.s. this is bad idea...
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

		go func() {
			for {
				ws.WriteMessage(websocket.PingMessage, []byte{})
				time.Sleep(time.Second * 10)
			}
		}()

		// don't forget to close the connection & remove client
		defer ws.Close()
		defer delete(controller.clients, ws)

		// first message on WS should be our hello
		var deviceInfo struct {
			Serial string `json:"serial"`
		}
		err := ws.ReadJSON(&deviceInfo)

		if err != nil {
			log.Errorf("I don't like this WS Client: %v", err)
			ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "I expect you to behave nicely"))
			return
		} else {
			log.Infof("[%v] connected from: %v", deviceInfo.Serial, ws.RemoteAddr())

			device := remote.RemoteServerClient{
				Address: ws.RemoteAddr().String(),
				Name:    deviceInfo.Serial,
			}

			// register handler && start the loop
			handler := remote.NewRemoteServerWS(&device, ws, controller.templatesDir, controller.tessdataDir)

			// put the client into active clients...
			controller.clients[ws] = device

			// handle the command / send actions
			handler.Loop()
		}
	})
}
