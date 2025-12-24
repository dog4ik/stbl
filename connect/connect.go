package connect

type GwConnectError struct {
	Result bool             `json:"result"`
	Error  string           `json:"error"`
	Logs   []InteractionLog `json:"logs"`
}
