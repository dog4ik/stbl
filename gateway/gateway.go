package gateway

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dog4ik/stbl/connect"
	"github.com/dog4ik/stbl/db"
	"github.com/dog4ik/stbl/utils"
)

const ACCESS_TOKEN_TTL = 15 * time.Minute

type GatewayError struct {
	Detail *string `json:"detail"`
}

type GatewayClient struct {
	client       *http.Client
	refreshToken string
	accessToken  string
	baseUrl      string
	callbackUrl  string
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func obtainFreshTokens(client *http.Client, il *connect.InteractionLogs, baseURL, login, password string) (*AuthResponse, error) {
	authReq := AuthRequest{
		Username: login,
		Password: password,
	}
	jsonData, err := json.Marshal(authReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	logger := il.Enter("login")
	url := baseURL + "/auth/api/v1/external-tokens/token-obtain"
	logger.SetRequest(utils.SecureStruct(authReq), url)

	res, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("POST request failed: %v", err)
	}
	defer res.Body.Close()
	logger.SetStatus(res.StatusCode)

	if res.StatusCode != http.StatusCreated {
		utils.DecodeBody(res.Body, logger)
		return nil, fmt.Errorf("unexpected status: %s", res.Status)
	}

	body := utils.DecodeBody(res.Body, logger)
	var authRes AuthResponse
	if err := json.Unmarshal(body, &authRes); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	return &authRes, nil
}

func refreshAccessToken(client *http.Client, il *connect.InteractionLogs, baseUrl, refreshToken string) (*AuthResponse, error) {
	reqBody := RefreshRequest{
		RefreshToken: refreshToken,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	logger := il.Enter("refresh_token")
	url := baseUrl + "/auth/api/v1/external-tokens/token-refresh"
	logger.SetRequest(string(jsonData), url)

	res, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("POST request failed: %v", err)
	}
	defer res.Body.Close()
	logger.SetStatus(res.StatusCode)

	if res.StatusCode != http.StatusCreated {
		utils.DecodeBody(res.Body, logger)
		return nil, fmt.Errorf("unexpected status: %s", res.Status)
	}

	body := utils.DecodeBody(res.Body, logger)
	var authRes AuthResponse
	if err := json.Unmarshal(body, &authRes); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	return &authRes, nil
}

func NewGatewayClient(
	ctx context.Context,
	settings connect.Settings,
	client *http.Client,
	conn *db.Queries,
	prodBaseUrl, sandoxBaseUrl, callbackUrl string,
) (GatewayClient, connect.InteractionLogs, error) {
	var baseUrl string
	if settings.Sandbox {
		baseUrl = sandoxBaseUrl
	} else {
		baseUrl = prodBaseUrl
	}

	il := connect.EmptyInteractionLogs()

	// BAD, but I log settings anyway :3
	sum := sha256.Sum256(fmt.Appendf(nil, "%s:%s", settings.Login, settings.Password))
	credentialsHash := hex.EncodeToString(sum[:])

	cached, err := conn.GetTokenCache(ctx, credentialsHash)
	if err == nil {
		// "Перед каждым запросом всегда проверяйте время жизни токена. Если до его истечения остается менее минуты, получите новый токен доступа к сервису."
		if time.Since(cached.AccessRefreshedAt) < ACCESS_TOKEN_TTL-time.Minute {
			log.Printf("Using cached access token: %s", cached.AccessToken)
			return GatewayClient{
				client:       client,
				accessToken:  cached.AccessToken,
				refreshToken: cached.RefreshToken,
				baseUrl:      baseUrl,
				callbackUrl:  callbackUrl,
			}, il, nil
		}

		log.Printf("Refreshing expired access token with refresh token: %s", cached.RefreshToken)
		refreshRes, err := refreshAccessToken(
			client,
			&il,
			baseUrl,
			cached.RefreshToken,
		)
		if err == nil {
			_ = conn.UpsertTokenCache(ctx, db.UpsertTokenCacheParams{
				CredentialsHash: credentialsHash,
				AccessToken:     refreshRes.AccessToken,
				RefreshToken:    cached.RefreshToken,
			})
			return GatewayClient{
				client:       client,
				accessToken:  refreshRes.AccessToken,
				refreshToken: cached.RefreshToken,
				baseUrl:      baseUrl,
				callbackUrl:  callbackUrl,
			}, il, nil
		} else {
			log.Printf("Failed to refresh access token: %v", err)
		}
	} else {
		log.Printf("Failed fetch auth tokens from db: %v", err)
	}

	log.Printf("Obtaining fresh pair of access and refresh tokens")
	auth, err := obtainFreshTokens(
		client,
		&il,
		baseUrl,
		settings.Login,
		settings.Password,
	)
	if err != nil {
		return GatewayClient{}, il, fmt.Errorf("failed to login client: %w", err)
	}

	conn.UpsertTokenCache(ctx, db.UpsertTokenCacheParams{
		CredentialsHash: credentialsHash,
		AccessToken:     auth.AccessToken,
		RefreshToken:    auth.RefreshToken,
	})

	return GatewayClient{
		client:       client,
		refreshToken: auth.RefreshToken,
		accessToken:  auth.AccessToken,
		baseUrl:      baseUrl,
		callbackUrl:  callbackUrl,
	}, il, nil
}

func (self *GatewayClient) makeRequest(method string, path string, body any, logger *connect.LogWriter) (*http.Response, error) {
	url := self.baseUrl + path
	log.Printf("DEBUG: Making %s request to %s", method, url)

	var (
		bodyBytes []byte
		err       error
	)

	if body != nil {
		securedBody := utils.SecureStruct(body)
		log.Printf("DEBUG: Gateway request body: %s", securedBody)

		if logger != nil {
			logger.SetRequest(securedBody, url)
		}

		bodyBytes, err = json.Marshal(body)

		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	} else if logger != nil {
		logger.SetRequest("", url)
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+self.accessToken)

	res, err := self.client.Do(req)
	if err != nil {
		return nil, err
	}

	log.Printf("Gateway response status: %s", res.Status)
	if logger != nil {
		logger.SetStatus(res.StatusCode)
	}
	return res, nil
}

func (self *GatewayClient) Payment(req connect.PayoutRequest, logger *connect.LogWriter) (*http.Response, error) {
	if req.Payment.GatewayAmount == nil || req.Payment.GatewayCurrency == nil {
		return nil, fmt.Errorf("Gateway connect request missing required fields")
	}

	paymentRequest := PaymentRequest{
		Amount:         float64(*req.Payment.GatewayAmount) / 100.0,
		TransferMethod: "QR_CODE",
		BankName:       "",
		ExternalID:     req.Payment.Token,
		AdditionalData: PaymentAdditionalData{},
		ClientID:       "",
	}

	return self.makeRequest(http.MethodPost, "/pay/external-api/v1/payments", paymentRequest, logger)
}

func (self *GatewayClient) Payout(req connect.PayoutRequest, logger *connect.LogWriter) (*http.Response, error) {
	if req.Payment.GatewayAmount == nil || req.Params.BankAccount.AccountNumber == nil {
		return nil, fmt.Errorf("Gateway connect request missing required fields")
	}

	payoutRequest := PayoutRequest{
		Amount:         float64(*req.Payment.GatewayAmount) / 100.0,
		TransferMethod: "CBU",
		BankCardNumber: "",
		PhoneNumber:    "",
		BankName:       "",
		AdditionalData: &PayoutAdditionalData{
			CBU: *req.Params.BankAccount.AccountNumber,
		},
		ExternalID: req.Payment.Token,
	}

	return self.makeRequest(http.MethodPost, "/pay/external-api/v1/payouts", payoutRequest, logger)
}

func (self *GatewayClient) RequestPaymentStatus(req connect.StatusRequest, logger *connect.LogWriter) (*http.Response, error) {
	if req.Payment.GatewayToken == nil {
		return nil, fmt.Errorf("Gateway connect request is missing required fields")
	}
	return self.makeRequest(http.MethodGet, "/pay/external-api/v1/payments/"+*req.Payment.GatewayToken, nil, logger)
}

func (self *GatewayClient) RequestPayoutStatus(req connect.StatusRequest, logger *connect.LogWriter) (*http.Response, error) {
	if req.Payment.GatewayToken == nil {
		return nil, fmt.Errorf("Gateway connect request is missing required fields")
	}
	return self.makeRequest(http.MethodGet, "/pay/external-api/v1/payouts/"+*req.Payment.GatewayToken, nil, logger)
}
