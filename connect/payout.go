package connect

type Params struct {
	Customer    Customer     `json:"customer"`
	Card        Card         `json:"card"`
	BankAccount *BankAccount `json:"bank_account"`
}

type BankAccount struct {
	RequisiteType string  `json:"requisite_type"`
	BankName      *string `json:"bank_name"`
	AccountNumber *string `json:"account_number"`
}

type Customer struct {
	Ip         *string `json:"ip"`
	FirstName  *string `json:"first_name"`
	MiddleName *string `json:"middle_name"`
	LastName   *string `json:"last_name"`
	Email      string  `json:"email"`
	Phone      string  `json:"phone"`
}

func (self *Customer) MakeFio() string {
	out := ""
	firstName, middleName, lastName := self.FirstName, self.MiddleName, self.LastName
	if firstName != nil && len(*firstName) != 0 {
		out = *firstName
	}

	if middleName != nil && len(*middleName) != 0 {
		if out != "" {
			out += " "
		}
		out += *middleName
	}

	if lastName != nil && len(*lastName) != 0 {
		if out != "" {
			out += " "
		}
		out += *lastName
	}

	return out
}

type Card struct {
	Pan string `json:"pan"`
}

type Payment struct {
	Token              string  `json:"token"`
	CallbackURL        string  `json:"callback_url"`
	MerchantPrivateKey string  `json:"merchant_private_key"`
	ExtraReturnParam   *string `json:"extra_return_param"`
	GatewayCurrency    *string `json:"gateway_currency"`
	GatewayAmount      *int    `json:"gateway_amount"`
	LeadId             int     `json:"lead_id"`
}

type PayoutRequest struct {
	Params        Params   `json:"params"`
	Payment       Payment  `json:"payment"`
	ProcessingUrl string   `json:"processing_url"`
	Settings      Settings `json:"settings"`
}

type RequisiteDetails struct {
	Bank   *string
	Number string
	Method string
}

type PayoutResponse struct {
	Result          bool             `json:"result"`
	Logs            []InteractionLog `json:"logs"`
	RedirectRequest RedirectRequest  `json:"redirect_request"`
	Status          string           `json:"status"`
	GatewayToken    *string          `json:"gateway_token,omitempty"`
}

type RedirectRequest struct {
	URL  string              `json:"url"`
	Type RedirectRequestType `json:"type"`
}

func NewGetRedirect(url string) RedirectRequest {
	return RedirectRequest{URL: url, Type: RedirectGetWithProcessing}
}

type RedirectRequestType string

const (
	RedirectPostIframes       RedirectRequestType = "post_iframes"
	RedirectGetWithProcessing RedirectRequestType = "get_with_processing"
	RedirectGet               RedirectRequestType = "get"
	RedirectPost              RedirectRequestType = "post"
	RedirectRedirectHtml      RedirectRequestType = "redirect_html"
)
