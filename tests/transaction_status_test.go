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

func TestGetTransactionStatus_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 1
	now := time.Now()
	expectedTx := db.Transaction{
		ID:        txID,
		UserID:    1,
		Amount:    decimal.NewFromFloat(100.0),
		Currency:  "USD",
		Type:      "deposit",
		Status:    "completed",
		GatewayID: 1,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now,
	}

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(expectedTx, nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	tx, err := service.GetTransactionStatus(ctx, txID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTx.ID, tx.ID)
	assert.Equal(t, expectedTx.Status, tx.Status)
}

func TestGetTransactionStatus_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 999 // Non-existent transaction

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(db.Transaction{}, fmt.Errorf("transaction not found"))

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	_, err := service.GetTransactionStatus(ctx, txID)

	assert.Error(t, err)
}

func TestGetTransactionStatus_WithGatewayInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 1
	gatewayTxnID := "gateway-txn-123"
	now := time.Now()
	expectedTx := db.Transaction{
		ID:           txID,
		UserID:       1,
		Amount:       decimal.NewFromFloat(100.0),
		Currency:     "USD",
		Type:         "deposit",
		Status:       "completed",
		GatewayID:    2, // Different gateway than originally assigned
		GatewayTxnID: gatewayTxnID,
		CreatedAt:    now.Add(-time.Hour),
		UpdatedAt:    now,
	}

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(expectedTx, nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	tx, err := service.GetTransactionStatus(ctx, txID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTx.ID, tx.ID)
	assert.Equal(t, expectedTx.Status, tx.Status)
	assert.Equal(t, expectedTx.GatewayID, tx.GatewayID)
	assert.Equal(t, expectedTx.GatewayTxnID, tx.GatewayTxnID)
}

func TestGetTransactionStatus_FailedTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 1
	now := time.Now()
	expectedTx := db.Transaction{
		ID:           txID,
		UserID:       1,
		Amount:       decimal.NewFromFloat(100.0),
		Currency:     "USD",
		Type:         "deposit",
		Status:       "failed",
		GatewayID:    1,
		ErrorMessage: "All payment gateways failed",
		CreatedAt:    now.Add(-time.Hour),
		UpdatedAt:    now,
	}

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(expectedTx, nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	tx, err := service.GetTransactionStatus(ctx, txID)

	assert.NoError(t, err)
	assert.Equal(t, expectedTx.ID, tx.ID)
	assert.Equal(t, "failed", tx.Status)
	assert.Equal(t, expectedTx.ErrorMessage, tx.ErrorMessage)
}

func TestGetTransactionStatus_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	txID := 1

	mockDB.EXPECT().GetTransactionByID(gomock.Any(), txID).Return(db.Transaction{}, fmt.Errorf("database connection error"))

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	_, err := service.GetTransactionStatus(ctx, txID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection error")
}
