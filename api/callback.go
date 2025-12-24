package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dog4ik/stbl/connect"
)

func callbackError(w http.ResponseWriter, msg string, err error) {
	log.Printf("ERROR: %s: %v", msg, err)
	http.Error(w, msg, http.StatusBadRequest)
}

type gatewayCallbackParams struct {
	gatewayID string
	Status    string
	Amount    int
	Reason    *string
	Token     string
}

func (state *ApiState) sendGatewayCallback(
	w http.ResponseWriter,
	r *http.Request,
	params gatewayCallbackParams,
) {
	mapping, err := state.queries.GetMapping(r.Context(), params.gatewayID)
	if err != nil {
		callbackError(w, "failed to load gateway token mapping", err)
		return
	}

	payload := connect.CallbackPayload{
		Currency: "ARS",
		Status:   params.Status,
		Amount:   params.Amount,
		Reason:   params.Reason,
	}

	jwt, err := connect.CreateJWT(payload, mapping.MerchantPrivateKey, []byte(state.signKey))
	if err != nil {
		callbackError(w, "failed to create JWT", err)
		return
	}

	jsonPayload, _ := json.Marshal(payload)

	url := fmt.Sprintf(
		"%s/callbacks/v2/gateway_callbacks/%s",
		state.businessUrl,
		params.Token,
	)

	log.Printf("Sending gateway connect callback(%s) payload: %s", url, jsonPayload)

	req, err := http.NewRequestWithContext(
		r.Context(),
		http.MethodPost,
		url,
		bytes.NewReader(jsonPayload),
	)
	if err != nil {
		callbackError(w, "failed to create callback request", err)
		return
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "Bearer "+jwt)

	res, err := state.client.Do(req)
	if err != nil {
		callbackError(w, "failed to send callback", err)
		return
	}
	defer res.Body.Close()

	log.Printf("Gateway connect callback response: %s", res.Status)
}
