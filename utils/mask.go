package utils

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
)

func mask(card string) string {
	length := len(card)
	if length > 10 {
		maskedMiddle := strings.Repeat("*", length-10)
		return card[:6] + maskedMiddle + card[length-4:]
	}
	return card
}

func isPANKey(key string) bool {
	k := strings.ToLower(key)
	switch k {
	case "pan", "cbu", "cbui", "number":
		return true
	}
	return false
}

func isCVVKey(key string) bool {
	k := strings.ToLower(key)
	return strings.Contains(k, "cvv") ||
		strings.Contains(k, "cvc") ||
		strings.Contains(k, "card_verification") ||
		strings.Contains(k, "cvn")
}

func SecureJSON(data any) string {
	secured := secureValue(data)

	result, err := json.Marshal(secured)
	if err != nil {
		log.Printf("ERROR: failed to secure json payload: %s", err)
		return ""
	}

	return string(result)
}

// secureValue recursively walks the JSON structure and masks PAN/CVV fields
func secureValue(v any) any {
	switch value := v.(type) {
	case map[string]any:
		newMap := make(map[string]any, len(value))
		for k, val := range value {
			isPan := isPANKey(k)
			isCvv := isCVVKey(k)

			newVal := val
			if isPan || isCvv {
				switch typed := val.(type) {
				case string:
					if isPan {
						newVal = mask(typed)
					} else if isCvv {
						newVal = "***"
					}
				// "UseNumber causes the Decoder to unmarshal a number into an interface value as a Number instead of as a float64."
				// and we don't use it here so handle float64
				case float64:
					str := strconv.FormatFloat(typed, 'f', 0, 64)
					if isPan {
						newVal = mask(str)
					} else if isCvv {
						newVal = "***"
					}
				case json.Number:
					str := typed.String()
					if isPan {
						newVal = mask(str)
					} else if isCvv {
						newVal = "***"
					}
				}
			} else {
				newVal = secureValue(val)
			}
			newMap[k] = newVal
		}
		return newMap

	case []any:
		newArr := make([]any, len(value))
		for i, item := range value {
			newArr[i] = secureValue(item)
		}
		return newArr

	default:
		return v
	}
}

func SecureStruct[T any](in T) string {
	b, err := json.Marshal(in)
	if err != nil {
		return ""
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return ""
	}

	secured := secureValue(m)

	b2, err := json.Marshal(secured)
	if err != nil {
		return ""
	}

	return string(b2)

}
