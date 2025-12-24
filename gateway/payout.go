package gateway

import "log"

type PayoutRequest struct {
	Amount         float64 `json:"amount"`
	TransferMethod string  `json:"transfer_method,omitempty"`

	BankCardNumber string `json:"bank_card_number,omitempty"`

	PhoneNumber string `json:"phone_number,omitempty"`
	BankName    string `json:"bank_name,omitempty"`

	AdditionalData *PayoutAdditionalData `json:"additional_data,omitempty"`

	ExternalID string `json:"external_id,omitempty"`
}

type PayoutAdditionalData struct {
	CustomerID string `json:"customer_id,omitempty"`
	FullName   string `json:"full_name,omitempty"`
	CUIT       string `json:"cuit,omitempty"`
	CBU        string `json:"cbu,omitempty"`
}

type PayoutResponse struct {
	ID             *string      `json:"id"`
	Num            string       `json:"num"`
	Amount         float64      `json:"amount"`
	BankCardNumber string       `json:"bank_card_number"`
	PhoneNumber    string       `json:"phone_number"`
	CreatedAt      string       `json:"created_at"`
	UpdatedAt      string       `json:"updated_at"`
	Status         PayoutStatus `json:"status"`
	ExternalID     string       `json:"external_id"`
	BankName       string       `json:"bank_name"`
}

type PayoutStatus struct {
	Name      StblPayoutStatus `json:"name"`
	UpdatedAt string           `json:"updated_at"`
}

type PayoutStatusResponse struct {
	ID             *string      `json:"id"`
	Num            string       `json:"num"`
	Amount         *float64     `json:"amount"`
	BankCardNumber string       `json:"bank_card_number"`
	PhoneNumber    string       `json:"phone_number"`
	CreatedAt      string       `json:"created_at"`
	UpdatedAt      string       `json:"updated_at"`
	Status         PayoutStatus `json:"status"`
	ExternalID     string       `json:"external_id"`
	BankName       string       `json:"bank_name"`
}

type StblPayoutStatus string

const (
	PayoutStatusAwaitingProcessing   StblPayoutStatus = "AWAITING_PROCESSING"
	PayoutStatusAwaitingConfirmation StblPayoutStatus = "AWAITING_CONFIRMATION"
	PayoutStatusDenied               StblPayoutStatus = "PAYOUT_DENIED"
	PayoutStatusPaid                 StblPayoutStatus = "PAID"
)

func (status StblPayoutStatus) ToRPStatus() string {
	switch status {
	case PayoutStatusAwaitingConfirmation, PayoutStatusAwaitingProcessing:
		return "pending"
	case PayoutStatusDenied:
		return "declined"
	case PayoutStatusPaid:
		return "approved"
	default:
		log.Printf("WARN: unhandled payout status: %s", status)
		return "pending"
	}
}
