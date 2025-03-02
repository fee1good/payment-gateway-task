package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"payment-gateway/configs/envs"
	"payment-gateway/tests/mocks"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"payment-gateway/db"
	"payment-gateway/internal/services"
)

func TestHandleCallback_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 1
	gatewayTxnID := "gateway-txn-1"
	status := "success"
	now := time.Now()
	tx := db.Transaction{
		ID:        txID,
		UserID:    1,
		Amount:    decimal.NewFromFloat(100.0),
		Currency:  "USD",
		Type:      "deposit",
		Status:    "processing",
		GatewayID: 1,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	}

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(tx, nil)
	mockDB.EXPECT().UpdateTransactionStatus(gomock.Any(), txID, "completed", gatewayTxnID, "").Return(nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.HandleCallback(ctx, gatewayTxnID, status, txID)

	assert.NoError(t, err)
}

func TestHandleCallback_TransactionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 999
	gatewayTxnID := "gateway-txn-1"
	status := "success"

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(db.Transaction{}, assert.AnError)

	cfg := &envs.Config{}
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.HandleCallback(ctx, gatewayTxnID, status, txID)

	assert.Error(t, err)
}

func TestHandleCallback_FailedStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 1
	gatewayTxnID := "gateway-txn-1"
	status := "failed"
	now := time.Now()
	tx := db.Transaction{
		ID:        txID,
		UserID:    1,
		Amount:    decimal.NewFromFloat(100.0),
		Currency:  "USD",
		Type:      "deposit",
		Status:    "processing",
		GatewayID: 1,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	}

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(tx, nil)
	mockDB.EXPECT().UpdateTransactionStatus(gomock.Any(), txID, "failed", gatewayTxnID, "").Return(nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.HandleCallback(ctx, gatewayTxnID, status, txID)

	assert.NoError(t, err)
}

func TestHandleCallback_AlreadyInFinalState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 1
	gatewayTxnID := "gateway-txn-1"
	status := "success"
	now := time.Now()
	tx := db.Transaction{
		ID:        txID,
		UserID:    1,
		Amount:    decimal.NewFromFloat(100.0),
		Currency:  "USD",
		Type:      "deposit",
		Status:    "completed", // Already in final state
		GatewayID: 1,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	}

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(tx, nil)
	// No update should be called

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.HandleCallback(ctx, gatewayTxnID, status, txID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in final state")
}

func TestHandleCallback_UpdateTransactionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 1
	gatewayTxnID := "gateway-txn-1"
	status := "success"
	now := time.Now()
	tx := db.Transaction{
		ID:        txID,
		UserID:    1,
		Amount:    decimal.NewFromFloat(100.0),
		Currency:  "USD",
		Type:      "deposit",
		Status:    "processing",
		GatewayID: 1,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	}

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(tx, nil)
	mockDB.EXPECT().UpdateTransactionStatus(gomock.Any(), txID, "completed", gatewayTxnID, "").
		Return(fmt.Errorf("database error"))

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.HandleCallback(ctx, gatewayTxnID, status, txID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update transaction status")
}

func TestHandleCallback_WithFallbackGateway(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 1
	gatewayTxnID := "fallback-gateway-txn-1"
	status := "success"
	now := time.Now()
	tx := db.Transaction{
		ID:        txID,
		UserID:    1,
		Amount:    decimal.NewFromFloat(100.0),
		Currency:  "USD",
		Type:      "deposit",
		Status:    "processing",
		GatewayID: 2, // This is a fallback gateway ID (not the original one)
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	}

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(tx, nil)
	mockDB.EXPECT().UpdateTransactionStatus(gomock.Any(), txID, "completed", gatewayTxnID, "").Return(nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.HandleCallback(ctx, gatewayTxnID, status, txID)

	assert.NoError(t, err)
}

func TestHandleCallback_WithCustomStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 1
	gatewayTxnID := "gateway-txn-1"
	status := "approved" // Different status string that should map to "completed"
	now := time.Now()
	tx := db.Transaction{
		ID:        txID,
		UserID:    1,
		Amount:    decimal.NewFromFloat(100.0),
		Currency:  "USD",
		Type:      "deposit",
		Status:    "processing",
		GatewayID: 1,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	}

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(tx, nil)
	mockDB.EXPECT().UpdateTransactionStatus(gomock.Any(), txID, "completed", gatewayTxnID, "").Return(nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.HandleCallback(ctx, gatewayTxnID, status, txID)

	assert.NoError(t, err)
}

func TestHandleCallback_WithRejectedStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 1
	gatewayTxnID := "gateway-txn-1"
	status := "rejected" // Different status string that should map to "failed"
	now := time.Now()
	tx := db.Transaction{
		ID:        txID,
		UserID:    1,
		Amount:    decimal.NewFromFloat(100.0),
		Currency:  "USD",
		Type:      "deposit",
		Status:    "processing",
		GatewayID: 1,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	}

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(tx, nil)
	mockDB.EXPECT().UpdateTransactionStatus(gomock.Any(), txID, "failed", gatewayTxnID, "").Return(nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.HandleCallback(ctx, gatewayTxnID, status, txID)

	assert.NoError(t, err)
}
