package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dog4ik/stbl/connect"
	"github.com/dog4ik/stbl/db"
	"github.com/dog4ik/stbl/gateway"
	"github.com/dog4ik/stbl/utils"
)

type ApiState struct {
	client            *http.Client
	queries           *db.Queries
	businessUrl       string
	signKey           string
	sandboxGatewayUrl string
	prodGatewayUrl    string
	callbackUrl       string
}

func NewState(queries *db.Queries) *ApiState {
	businessUrl := utils.ExpectEnv("BUSINESS_URL")
	signKey := utils.ExpectEnv("SIGN_KEY")
	sandboxGatewayUrl := utils.ExpectEnv("SANDBOX_BASE_URL")
	prodGatewayUrl := utils.ExpectEnv("BASE_URL")
	client := &http.Client{Timeout: 30 * time.Second}

	return &ApiState{
		client:            client,
		queries:           queries,
		businessUrl:       businessUrl,
		signKey:           signKey,
		sandboxGatewayUrl: sandboxGatewayUrl,
		prodGatewayUrl:    prodGatewayUrl,
	}
}

func (state *ApiState) newGatewayClient(ctx context.Context, settings connect.Settings) (gateway.GatewayClient, connect.InteractionLogs, error) {
	return gateway.NewGatewayClient(ctx, settings, state.client, state.queries, state.prodGatewayUrl, state.sandboxGatewayUrl, state.callbackUrl)
}

func writePayoutPendingResponse(w http.ResponseWriter, interactionLogs connect.InteractionLogs, redirect connect.RedirectRequest) {
	utils.WriteJSON(
		w,
		connect.PayoutResponse{
			Result:          true,
			Logs:            interactionLogs.IntoInner(),
			RedirectRequest: redirect,
			Status:          "pending",
			GatewayToken:    nil,
		},
	)
}

func writeErrorResponse(w http.ResponseWriter, interactionLogs connect.InteractionLogs, msg string) {
	utils.WriteJSON(
		w,
		connect.GwConnectError{
			Result: false,
			Logs:   interactionLogs.IntoInner(),
			Error:  msg,
		},
	)
}

func gatewayErrorMessage(body []byte) string {
	var ge gateway.GatewayError
	if err := json.Unmarshal(body, &ge); err == nil && ge.Detail != nil {
		return *ge.Detail
	}
	return "bad gateway response"
}

func (state *ApiState) PaymentHandler(w http.ResponseWriter, r *http.Request) {
	payment, err := utils.DecodeJSONRequest[connect.PayoutRequest](r.Body, w)
	if err != nil {
		log.Printf("Failed to decode gateway connect request: %s", err)
		writeErrorResponse(w, connect.EmptyInteractionLogs(), err.Error())
		return
	}

	client, il, err := state.newGatewayClient(r.Context(), payment.Settings)
	if err != nil {
		log.Printf("Failed to initiate gateway client: %v", err)
		writeErrorResponse(w, il, err.Error())
		return
	}

	span := il.Enter("payment")
	res, err := client.Payment(payment, span)
	if err != nil {
		log.Printf("ERROR: Failed to create payout: %s", err)
		writeErrorResponse(w, il, fmt.Sprintf("Gateway request failed: %s", err))
		return
	}
	defer res.Body.Close()

	body := utils.DecodeBody(res.Body, span)
	if res.StatusCode == http.StatusCreated {
		gatewayPayment, err := utils.UnmarshalBytes[gateway.PaymentResponse](body)
		// json deserialization error
		if err != nil {
			writeErrorResponse(w, il, fmt.Sprintf("Failed to deserilaize gateway response: %s", err))
			return
		}

		// required fields are missing
		if gatewayPayment.ID == nil {
			writeErrorResponse(w, il, fmt.Sprintf("Payment response missing required fields: %s", err))
			return
		}

		if _, err = state.queries.CreateMapping(
			r.Context(),
			db.CreateMappingParams{Token: payment.Payment.Token, MerchantPrivateKey: payment.Payment.MerchantPrivateKey, GatewayID: *gatewayPayment.ID},
		); err != nil {
			log.Printf("ERROR: Failed to insert gateway token mapping: %s", err)
		}

		utils.WriteJSON(
			w,
			connect.PayoutResponse{
				Result:          true,
				Logs:            il.IntoInner(),
				RedirectRequest: connect.NewGetRedirect(gatewayPayment.PayFormLink),
				Status:          gatewayPayment.Status.Name.ToRPStatus(),
				GatewayToken:    gatewayPayment.ID,
			},
		)
	} else {
		utils.WriteJSON(
			w,
			connect.GwConnectError{
				Result: false,
				Error:  gatewayErrorMessage(body),
				Logs:   il.IntoInner(),
			},
		)
	}
}
func (state *ApiState) PayoutHandler(w http.ResponseWriter, r *http.Request) {
	payout, err := utils.DecodeJSONRequest[connect.PayoutRequest](r.Body, w)
	if err != nil {
		log.Printf("Failed to decode gateway connect request: %s", err)
		writeErrorResponse(w, connect.EmptyInteractionLogs(), err.Error())
		return
	}

	client, il, err := state.newGatewayClient(r.Context(), payout.Settings)
	if err != nil {
		log.Printf("Failed to initiate gateway client: %v", err)
		writeErrorResponse(w, il, err.Error())
		return
	}

	span := il.Enter("payout")

	res, err := client.Payout(payout, span)
	if err != nil {
		log.Printf("ERROR: Failed to create payout: %s", err)
		writeErrorResponse(w, il, fmt.Sprintf("Gateway request failed: %s", err))
		return
	}
	defer res.Body.Close()

	body := utils.DecodeBody(res.Body, span)
	if res.StatusCode == http.StatusCreated {
		providerPayout, err := utils.UnmarshalBytes[gateway.PayoutResponse](body)
		// json deserialization error
		if err != nil {
			writePayoutPendingResponse(w, il, connect.NewGetRedirect(payout.ProcessingUrl))
			return
		}

		// required fields are missing
		if providerPayout.ID == nil {
			writePayoutPendingResponse(w, il, connect.NewGetRedirect(payout.ProcessingUrl))
			return
		}

		if _, err = state.queries.CreateMapping(
			r.Context(),
			db.CreateMappingParams{Token: payout.Payment.Token, MerchantPrivateKey: payout.Payment.MerchantPrivateKey, GatewayID: *providerPayout.ID},
		); err != nil {
			log.Printf("ERROR: Failed to insert gateway token mapping: %s", err)
		}

		utils.WriteJSON(
			w,
			connect.PayoutResponse{
				Result:          true,
				Logs:            il.IntoInner(),
				RedirectRequest: connect.NewGetRedirect(payout.ProcessingUrl),
				Status:          providerPayout.Status.Name.ToRPStatus(),
				GatewayToken:    providerPayout.ID,
			},
		)
	} else if res.StatusCode >= 500 {
		writePayoutPendingResponse(w, il, connect.NewGetRedirect(payout.ProcessingUrl))
	} else {
		var message string
		gatewayError, _ := utils.UnmarshalBytes[gateway.GatewayError](body)
		if gatewayError.Detail != nil {
			message = *gatewayError.Detail
		} else {
			writePayoutPendingResponse(w, il, connect.NewGetRedirect(payout.ProcessingUrl))
			return
		}

		utils.WriteJSON(
			w,
			connect.GwConnectError{
				Result: false,
				Error:  message,
				Logs:   il.IntoInner(),
			},
		)
	}
}

func (state *ApiState) StatusHandler(w http.ResponseWriter, r *http.Request) {
	status, err := utils.DecodeJSONRequest[connect.StatusRequest](r.Body, w)
	log.Printf("Status request: %v", utils.ToJSON(status))
	if err != nil {
		log.Printf("Failed to decode gateway connect request: %s", err)
		writeErrorResponse(w, connect.EmptyInteractionLogs(), err.Error())
		return
	}
	client, il, err := state.newGatewayClient(r.Context(), status.Settings)
	logger := il.Enter("status")
	if err != nil {
		log.Printf("Failed to initiate gateway client: %v", err)
		writeErrorResponse(w, il, err.Error())
		return
	}

	switch status.Payment.OperationType {
	case "pay":
		res, err := client.RequestPaymentStatus(status, logger)
		if err != nil {
			writeErrorResponse(w, il, err.Error())
			return
		}
		defer res.Body.Close()

		body := utils.DecodeBody(res.Body, logger)
		if res.StatusCode == http.StatusOK {
			providerStatus, err := utils.UnmarshalBytes[gateway.PaymentStatusResponse](body)
			// json deserialization error
			if err != nil {
				writeErrorResponse(w, il, err.Error())
				return
			}

			// missing required fields
			if providerStatus.ID == nil || providerStatus.Amount == nil {
				writeErrorResponse(w, il, "Incorrect provider response")
				return
			}

			utils.WriteJSON(
				w,
				connect.StatusResponse{
					Result: true,
					Logs:   il.IntoInner(),
					Status: providerStatus.Status.Name.ToRPStatus(),
					Amount: uint(*providerStatus.Amount * 100),
				},
			)
		} else {
			writeErrorResponse(w, il, gatewayErrorMessage(body))
		}
	case "payout":
		res, err := client.RequestPayoutStatus(status, logger)
		if err != nil {
			writeErrorResponse(w, il, err.Error())
			return
		}
		defer res.Body.Close()

		body := utils.DecodeBody(res.Body, logger)
		if res.StatusCode == http.StatusOK {
			providerStatus, err := utils.UnmarshalBytes[gateway.PayoutStatusResponse](body)
			// json deserialization error
			if err != nil {
				writeErrorResponse(w, il, err.Error())
				return
			}

			// missing required fields
			if providerStatus.ID == nil || providerStatus.Amount == nil {
				writeErrorResponse(w, il, "Incorrect provider response")
				return
			}

			utils.WriteJSON(
				w,
				connect.StatusResponse{
					Result: true,
					Logs:   il.IntoInner(),
					Status: providerStatus.Status.Name.ToRPStatus(),
					Amount: uint(*providerStatus.Amount * 100),
				},
			)
		} else {
			writeErrorResponse(w, il, gatewayErrorMessage(body))
		}
	default:
		log.Printf("WARN: Unsupported operation type: %s", status.Payment.OperationType)
	}
}

func (state *ApiState) PaymentCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received payment gateway callback")
	callback, err := utils.DecodeJSONRequest[gateway.PaymentCallback](r.Body, w)
	if err != nil {
		callbackError(w, "failed to decode callback body", err)
		return
	}

	if callback.ID == nil || callback.Amount == nil || callback.Status == nil {
		callbackError(w, "missing fields in gateway callback", fmt.Errorf("invalid payload"))
		return
	}

	mapping, err := state.queries.GetMapping(r.Context(), *callback.ID)
	if err != nil {
		callbackError(w, "failed to load gateway token mapping", err)
		return
	}

	amount := int(*callback.Amount * 100)

	if callback.NewAmount != nil {
		log.Printf("Got callback with updated amount: %.2f", *callback.NewAmount)
		amount = int(*callback.NewAmount * 100)
	}

	var reason *string
	if callback.Status.ToRPStatus() == "declined" {
		reason = (*string)(callback.Status)
	}

	state.sendGatewayCallback(w, r, gatewayCallbackParams{
		gatewayID: *callback.ID,
		Reason:    reason,
		Token:     mapping.Token,
		Status:    callback.Status.ToRPStatus(),
		Amount:    amount,
	})
}

func (state *ApiState) PayoutCallbackHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received payment gateway callback")
	callback, err := utils.DecodeJSONRequest[gateway.PayoutCallback](r.Body, w)
	if err != nil {
		callbackError(w, "failed to decode callback body", err)
		return
	}

	if callback.PayoutID == nil || callback.PayoutAmount == nil || callback.PayoutStatus == nil {
		callbackError(w, "missing fields in gateway callback", fmt.Errorf("invalid payload"))
		return
	}

	mapping, err := state.queries.GetMapping(r.Context(), *callback.PayoutID)
	if err != nil {
		callbackError(w, "failed to load gateway token mapping", err)
		return
	}

	amount := int(*callback.PayoutAmount * 100)

	if callback.PayoutAmount != nil {
		amount = int(*callback.PayoutAmount * 100)
	}

	var reason *string
	if callback.PayoutStatus.ToRPStatus() == "declined" {
		reason = (*string)(callback.PayoutStatus)
	}

	state.sendGatewayCallback(w, r, gatewayCallbackParams{
		gatewayID: *callback.PayoutID,
		Reason:    reason,
		Token:     mapping.Token,
		Status:    callback.PayoutStatus.ToRPStatus(),
		Amount:    amount,
	})
}
