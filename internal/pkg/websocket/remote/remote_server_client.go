package remote

// RemoteServerClient - holds basic information about rok-remote instance connected to websocket
type RemoteServerClient struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}
