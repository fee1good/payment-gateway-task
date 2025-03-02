package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"payment-gateway/configs/logger"
	"payment-gateway/db"
	"payment-gateway/internal/models"
	"payment-gateway/internal/services"

	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

type TransactionHandler struct {
	DB             db.Storage
	GatewayService services.GatewayServiceInterface
}

func NewTransactionHandler(db db.Storage, gatewayService services.GatewayServiceInterface) *TransactionHandler {
	return &TransactionHandler{
		DB:             db,
		GatewayService: gatewayService,
	}
}

func (h *TransactionHandler) DepositHandler(w http.ResponseWriter, r *http.Request) {
	var request models.TransactionRequest

	if err := DecodeRequest(r, &request); err != nil {
		logger.Error("Error decoding request", "error", err)
		response := models.APIResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request format",
		}
		err := EncodeResponse(w, r, response)
		if err != nil {
			logger.Warn("Error encoding response", "error", err)
			return
		}
		return
	}

	if request.Amount.LessThanOrEqual(decimal.Zero) {
		response := models.APIResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Amount must be greater than zero",
		}
		err := EncodeResponse(w, r, response)
		if err != nil {
			logger.Warn("Error encoding response", "error", err)
			return
		}
		return
	}

	transaction := db.Transaction{
		UserID:   request.UserID,
		Amount:   request.Amount,
		Currency: request.Currency,
		Type:     "deposit",
		Status:   "pending",
	}

	err := h.GatewayService.ProcessTransaction(r.Context(), transaction)
	if err != nil {
		logger.Error("Error processing deposit", "error", err)
		response := models.APIResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to process deposit",
		}
		err := EncodeResponse(w, r, response)
		if err != nil {
			logger.Warn("Error encoding response", "error", err)
			return
		}
		return
	}

	response := models.APIResponse{
		StatusCode: http.StatusAccepted,
		Message:    "Deposit request accepted and is being processed",
	}
	err = EncodeResponse(w, r, response)
	if err != nil {
		logger.Warn("Error encoding response", "error", err)
		return
	}
}

func (h *TransactionHandler) WithdrawalHandler(w http.ResponseWriter, r *http.Request) {
	var request models.TransactionRequest

	if err := DecodeRequest(r, &request); err != nil {
		logger.Error("Error decoding request", "error", err)
		response := models.APIResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request format",
		}
		err := EncodeResponse(w, r, response)
		if err != nil {
			logger.Warn("Error encoding response", "error", err)
			return
		}
		return
	}

	if request.Amount.LessThanOrEqual(decimal.Zero) {
		response := models.APIResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Amount must be greater than zero",
		}
		err := EncodeResponse(w, r, response)
		if err != nil {
			logger.Warn("Error encoding response", "error", err)
			return
		}
		return
	}

	transaction := db.Transaction{
		UserID:   request.UserID,
		Amount:   request.Amount,
		Currency: request.Currency,
		Type:     "withdrawal",
		Status:   "pending",
	}

	err := h.GatewayService.ProcessTransaction(r.Context(), transaction)
	if err != nil {
		logger.Error("Error processing withdrawal", "error", err)
		response := models.APIResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to process withdrawal",
		}
		err := EncodeResponse(w, r, response)
		if err != nil {
			logger.Warn("Error encoding response", "error", err)
			return
		}
		return
	}

	response := models.APIResponse{
		StatusCode: http.StatusAccepted,
		Message:    "Withdrawal request accepted and is being processed",
	}
	err = EncodeResponse(w, r, response)
	if err != nil {
		logger.Warn("Error encoding response", "error", err)
		return
	}
}

func (h *TransactionHandler) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transactionIDStr := vars["id"]
	transactionID, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		logger.Error("Invalid transaction ID", "error", err)
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var callbackData struct {
		GatewayTxnID string `json:"gateway_txn_id" xml:"gateway_txn_id"`
		Status       string `json:"status" xml:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&callbackData); err != nil {
		logger.Error("Error decoding callback data", "error", err)
		http.Error(w, "Invalid callback data", http.StatusBadRequest)
		return
	}

	err = h.GatewayService.HandleCallback(r.Context(), callbackData.GatewayTxnID, callbackData.Status, transactionID)
	if err != nil {
		logger.Error("Error processing callback", "error", err)
		http.Error(w, "Failed to process callback", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Callback processed successfully"))
	if err != nil {
		logger.Warn("Error writing response", "error", err)
		return
	}
}
