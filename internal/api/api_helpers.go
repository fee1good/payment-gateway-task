package api

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"

	"payment-gateway/internal/models"
)

func DecodeRequest(r *http.Request, request *models.TransactionRequest) error {
	contentType := r.Header.Get("Content-Type")

	switch contentType {
	case "application/json":
		return json.NewDecoder(r.Body).Decode(request)
	case "text/xml":
		return xml.NewDecoder(r.Body).Decode(request)
	case "application/xml":
		return xml.NewDecoder(r.Body).Decode(request)
	default:
		return fmt.Errorf("unsupported content type")
	}
}

func EncodeResponse(w http.ResponseWriter, r *http.Request, response models.APIResponse) error {
	contentType := r.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(response.StatusCode)

	switch contentType {
	case "application/json", "":
		// Default to JSON if no content type is specified
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(response)
	case "text/xml", "application/xml":
		return xml.NewEncoder(w).Encode(response)
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
}
