package wsserver

type wsMwssage struct {
	IPAddress string `json:"address"`
	Message   string `json:"message"`
	Time      string `json:"time"`
}
