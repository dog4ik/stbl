package gateway

type PaymentCallback struct {
	ID         *string            `json:"id"`
	Status     *StblPaymentStatus `json:"status"`
	Amount     *float64           `json:"amount"`
	ExternalID string             `json:"external_id"`
	NewAmount  *float64           `json:"new_amount"`
}

type PayoutCallback struct {
	PayoutID         *string           `json:"payout_id"`
	PayoutStatus     *StblPayoutStatus `json:"payout_status"`
	PayoutAmount     *float64          `json:"payout_amount"`
	PayoutExternalID string            `json:"payout_external_id"`
}
