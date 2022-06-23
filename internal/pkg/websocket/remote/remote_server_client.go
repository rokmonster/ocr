package remote

// ServerClient - holds basic information about rok-remote instance connected to websocket
type ServerClient struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}
