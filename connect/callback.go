package connect

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type CallbackPayload struct {
	Status   string  `json:"status"`
	Currency string  `json:"currency"`
	Amount   int     `json:"amount"`
	Reason   *string `json:"reason,omitempty"`
}

type SecureBlock struct {
	EncryptedData string `json:"encrypted_data"`
	IVValue       string `json:"iv_value"`
}

type JWTPayload struct {
	Payload CallbackPayload `json:"payload"`
	Secure  SecureBlock     `json:"secure"`
}

func genIV() []byte {
	iv := make([]byte, 16)
	rand.Read(iv)
	return iv
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := make([]byte, padding)
	for i := range padText {
		padText[i] = byte(padding)
	}
	return append(data, padText...)
}

// AES-256-CBC encryption with PKCS7 padding
func encryptMerchantKey(merchantKey string, signKey []byte, iv []byte) (string, string, error) {
	block, err := aes.NewCipher(signKey)
	if err != nil {
		return "", "", err
	}

	plainText := pkcs7Pad([]byte(merchantKey), block.BlockSize())
	mode := cipher.NewCBCEncrypter(block, iv)
	cipherText := make([]byte, len(plainText))
	mode.CryptBlocks(cipherText, plainText)

	encryptedData := base64.StdEncoding.EncodeToString(cipherText)
	ivValue := base64.StdEncoding.EncodeToString(iv)

	return encryptedData, ivValue, nil
}

// Encode JWT using HMAC-SHA512
func encodeJWT(payload any, signKey []byte) (string, error) {
	header := map[string]string{
		"alg": "HS512",
		"typ": "JWT",
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	base64Header := base64.RawURLEncoding.EncodeToString(headerJSON)
	base64Payload := base64.RawURLEncoding.EncodeToString(payloadJSON)

	message := fmt.Sprintf("%s.%s", base64Header, base64Payload)

	mac := hmac.New(sha512.New, signKey)
	mac.Write([]byte(message))
	signature := mac.Sum(nil)
	base64Signature := base64.RawURLEncoding.EncodeToString(signature)

	return fmt.Sprintf("%s.%s.%s", base64Header, base64Payload, base64Signature), nil
}

// Create JWT with encrypted merchant key
func CreateJWT(payload CallbackPayload, merchantKey string, signKey []byte) (string, error) {
	iv := genIV()

	encryptedData, ivValue, err := encryptMerchantKey(merchantKey, signKey, iv)
	if err != nil {
		return "", err
	}

	jwtPayload := JWTPayload{
		Payload: payload,
		Secure: SecureBlock{
			EncryptedData: encryptedData,
			IVValue:       ivValue,
		},
	}

	return encodeJWT(jwtPayload, signKey)
}
