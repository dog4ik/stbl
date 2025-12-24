package gateway

import "log"

type PaymentRequest struct {
	Amount         float64               `json:"amount,omitempty"`
	TransferMethod string                `json:"transfer_method,omitempty"`
	BankName       string                `json:"bank_name,omitempty"`
	ExternalID     string                `json:"external_id,omitempty"`
	AdditionalData PaymentAdditionalData `json:"additional_data"`
	ClientID       string                `json:"client_id,omitempty"`
}

type PaymentAdditionalData struct {
	FullName string `json:"full_name,omitempty"`
	CBU      string `json:"cbu,omitempty"`
	CUIT     string `json:"cuit,omitempty"`
}

type PaymentResponse struct {
	ID                 *string       `json:"id"`
	Amount             float64       `json:"amount"`
	BankName           string        `json:"bank_name"`
	TransferMethod     string        `json:"transfer_method"`
	BankCard           BankCard      `json:"bank_card"`
	ProviderPaymentURL string        `json:"provider_payment_url"`
	ProviderRequisite  string        `json:"provider_requisite"`
	Requisites         Requisites    `json:"requisites"`
	Status             PaymentStatus `json:"status"`
	PhoneNumber        string        `json:"phone_number"`
	ExternalID         string        `json:"external_id"`
	ExchangeRate       float64       `json:"exchange_rate"`
	PayFormLink        string        `json:"pay_form_link"`
}

type BankCard struct {
	QRCodeLink string `json:"qr_code_link"`
	FullName   string `json:"full_name"`
	Number     string `json:"number"`
}

type Requisites struct {
	CBU                  string `json:"cbu"`
	BoliviaAccountNumber string `json:"bolivia_account_number"`
	BoliviaQRCodeLink    string `json:"bolivia_qr_code_link"`
	ECUAccountNumber     string `json:"ecu_account_number"`
}

type PaymentStatus struct {
	Name      StblPaymentStatus `json:"name"`
	UpdatedAt string            `json:"updated_at"`
}

type PaymentStatusResponse struct {
	ID             *string       `json:"id"`
	Num            string        `json:"num"`
	Amount         *float64      `json:"amount"`
	TransferMethod string        `json:"transfer_method"`
	BankCard       BankCard      `json:"bank_card"`
	CreatedAt      string        `json:"created_at"`
	UpdatedAt      string        `json:"updated_at"`
	Status         PaymentStatus `json:"status"`
}

type StblPaymentStatus string

const (
	PayStatusNew                 StblPaymentStatus = "NEW"
	PayStatusCanceled            StblPaymentStatus = "CANCELED"
	PayStatusCompleted           StblPaymentStatus = "COMPLETED"
	PayStatusAppealApproved      StblPaymentStatus = "APPEAL_APPROVED"
	PayStatusAppealRejected      StblPaymentStatus = "APPEAL_REJECTED"
	PayStatusAppealConsideration StblPaymentStatus = "APPEAL_CONSIDERATION"
)

func (status StblPaymentStatus) ToRPStatus() string {
	switch status {
	case PayStatusNew, PayStatusAppealConsideration:
		return "pending"
	case PayStatusCompleted, PayStatusAppealApproved:
		return "approved"
	case PayStatusCanceled, PayStatusAppealRejected:
		return "declined"
	default:
		log.Printf("WARN: unhandled payment status: %s", status)
		return "pending"
	}
}
