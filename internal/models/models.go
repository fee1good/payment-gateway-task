package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type TransactionRequest struct {
	Amount   decimal.Decimal `json:"amount" xml:"amount"`
	UserID   int             `json:"user_id" xml:"user_id"`
	Currency string          `json:"currency" xml:"currency"`
}

type APIResponse struct {
	StatusCode int         `json:"status_code" xml:"status_code"`
	Message    string      `json:"message" xml:"message"`
	Data       interface{} `json:"data,omitempty" xml:"data,omitempty"`
}

type Transaction struct {
	ID           int             `json:"id"`
	UserID       int             `json:"user_id"`
	Amount       decimal.Decimal `json:"amount"`
	Currency     string          `json:"currency"`
	Type         string          `json:"type"`   // "deposit" or "withdrawal"
	Status       string          `json:"status"` // "pending", "completed", "failed"
	GatewayID    int             `json:"gateway_id"`
	GatewayTxnID string          `json:"gateway_txn_id,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
}
