package webcontrollers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	wsserver "github.com/rokmonster/ocr/internal/pkg/websocket/server"
	log "github.com/sirupsen/logrus"
)

func (controller *RemoteDevicesController) getRemoteDevices() map[uuid.UUID]ServerClient {
	return controller.clients
}

// ServerClient - holds basic information about rok-remote instance connected to websocket
type ServerClient struct {
	Name    string
	Address string
	Handler *wsserver.RemoteServerWS
}

type RemoteDevicesController struct {
	clients      map[uuid.UUID]ServerClient
	upgrader     websocket.Upgrader
	templatesDir string
	tessdataDir  string
}

func NewRemoteDevicesController(templates, tessdata string) *RemoteDevicesController {
	return &RemoteDevicesController{
		clients:      make(map[uuid.UUID]ServerClient),
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

func (controller *RemoteDevicesController) GetListOfDevices(c *gin.Context) {
	switch c.NegotiateFormat(gin.MIMEJSON, gin.MIMEHTML) {
	case gin.MIMEHTML:
		c.HTML(http.StatusOK, "devices.html", gin.H{
			"userdata": c.MustGet(AuthUserData),
			"devices":  controller.getRemoteDevices(),
		})
	case gin.MIMEJSON:
		c.JSON(http.StatusOK, gin.H{
			"devices": controller.getRemoteDevices(),
		})
	}
}

func (controller *RemoteDevicesController) Disconnect(ctx *gin.Context) {
	id := uuid.MustParse(ctx.Param("id"))
	if c, ok := controller.clients[id]; ok {
		c.Handler.Disconnect()
	}

	ctx.Redirect(http.StatusFound, "/devices/")
}

func (controller *RemoteDevicesController) Data(ctx *gin.Context) {
	id := uuid.MustParse(ctx.Param("id"))
	if c, ok := controller.clients[id]; ok {
		ctx.JSON(http.StatusOK, c.Handler.GetData())
		return
	}

	ctx.Redirect(http.StatusFound, "/devices/")

}

func (controller *RemoteDevicesController) Websocket(c *gin.Context) {
	ws, _ := controller.upgrader.Upgrade(c.Writer, c.Request, nil)

	// don't forget to close the connection & remove client
	defer ws.Close()

	// first message on WS should be our hello
	var deviceInfo struct {
		Serial string `json:"serial"`
	}
	err := ws.ReadJSON(&deviceInfo)

	if err != nil {
		log.Errorf("I don't like this WS Client: %v", err)
		_ = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "I expect you to behave nicely"))
		return
	} else {
		log.Infof("[%v] connected from: %v", deviceInfo.Serial, ws.RemoteAddr())

		sessionId := uuid.New()

		// register handler && start the loop
		handler := wsserver.NewRemoteServerWS(ws, sessionId.String(), controller.templatesDir, controller.tessdataDir)
		device := ServerClient{
			Address: ws.RemoteAddr().String(),
			Name:    deviceInfo.Serial,
			Handler: handler,
		}

		// put the client into active clients...
		controller.clients[sessionId] = device
		defer delete(controller.clients, sessionId)

		// handle the command / send actions
		handler.Loop()
	}
}
