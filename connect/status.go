package connect

type StatusRequest struct {
	Payment  StatusPayment `json:"payment"`
	Settings Settings      `json:"settings"`
}

type StatusPayment struct {
	GatewayToken  *string `json:"gateway_token,omitempty"`
	OperationType string  `json:"operation_type"`
	Token         string  `json:"token"`
}

type StatusResponse struct {
	Result   bool             `json:"result"`
	Logs     []InteractionLog `json:"logs,omitempty"`
	Status   string           `json:"status,omitempty"`
	Details  string           `json:"details,omitempty"`
	Amount   uint             `json:"amount,omitempty"`
	Currency string           `json:"currency,omitempty"`
}
