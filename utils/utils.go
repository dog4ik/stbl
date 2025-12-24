package utils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/dog4ik/stbl/connect"
)

func DecodeJSONRequest[T any](r io.Reader, w http.ResponseWriter) (T, error) {
	body := DecodeBody(r, nil)
	var v T

	v, err := UnmarshalBytes[T](body)
	if err != nil {
		log.Printf("ERROR: Error unmarshalling JSON: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return v, err
	}
	return v, nil
}

// Read the body and save it in interaction logs
func DecodeBody(r io.Reader, logger *connect.LogWriter) []byte {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return bytes
	}

	var temp any
	if err := json.Unmarshal(bytes, &temp); err != nil {
		log.Printf("ERROR: Body decoder failed to unmarshal JSON: %s", err)
		if logger != nil {
			logger.SetResponse(string(bytes))
		}
		return bytes
	}

	secured := SecureJSON(temp)
	log.Printf("DEBUG: JSON payload: %s", secured)
	if logger != nil {
		logger.SetResponse(secured)
	}
	return bytes
}

func UnmarshalBytes[T any](bytes []byte) (T, error) {
	var result T
	if err := json.Unmarshal(bytes, &result); err != nil {
		log.Printf("ERROR: Error converting JSON to target type: %s", err)
		return result, err
	}
	return result, nil
}

func DecodeJSONRespnose[T any](r *http.Response, logger *connect.LogWriter) (T, error) {
	var result T
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return result, err
	}
	var temp any
	if err := json.Unmarshal(bytes, &temp); err != nil {
		log.Printf("ERROR: Error unmarshalling JSON: %s", err)
		return result, err
	}

	secured := SecureJSON(temp)
	log.Printf("DEBUG: JSON response: %s", secured)
	if logger != nil {
		logger.SetResponse(secured)
	}

	if err := json.Unmarshal(bytes, &result); err != nil {
		log.Printf("ERROR: Error converting JSON to target type: %s", err)
		return result, err
	}

	return result, nil
}

func WriteJSON(w http.ResponseWriter, v any) {
	w.Header().Set("content-type", "application/json")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("ERROR: JSON encode failed: %v", err)
	}
}

func ToJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
