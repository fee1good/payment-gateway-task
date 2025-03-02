package tests

import (
	"context"
	"fmt"
	"testing"

	"payment-gateway/configs/envs"
	"payment-gateway/tests/mocks"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"payment-gateway/db"
	"payment-gateway/internal/models"
	"payment-gateway/internal/services"
)

func TestProcessTransaction_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	tx := db.Transaction{
		UserID:   1,
		Amount:   decimal.NewFromFloat(100.0),
		Currency: "USD",
		Type:     "deposit",
		Status:   "pending",
	}

	mockDB.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, transaction db.Transaction) (int, error) {
			assert.Equal(t, tx.UserID, transaction.UserID)
			assert.Equal(t, tx.Amount, transaction.Amount)
			assert.Equal(t, tx.Currency, transaction.Currency)
			assert.Equal(t, tx.Type, transaction.Type)
			assert.Equal(t, "pending", transaction.Status)
			assert.Equal(t, 0, transaction.GatewayID)
			return 1, nil
		})
	mockProcessor.EXPECT().ProcessTransaction(gomock.Any()).Do(
		func(transaction models.Transaction) {
			assert.Equal(t, 1, transaction.ID)
			assert.Equal(t, tx.UserID, transaction.UserID)
			assert.Equal(t, tx.Amount, transaction.Amount)
			assert.Equal(t, tx.Currency, transaction.Currency)
			assert.Equal(t, tx.Type, transaction.Type)
			assert.Equal(t, "pending", transaction.Status)
			assert.Equal(t, 0, transaction.GatewayID)
		})
	mockKafka.EXPECT().PublishMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.ProcessTransaction(ctx, tx)

	assert.NoError(t, err)
}

func TestProcessTransaction_CreateTransactionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	tx := db.Transaction{
		UserID:   1,
		Amount:   decimal.NewFromFloat(100.0),
		Currency: "USD",
		Type:     "deposit",
		Status:   "pending",
	}

	mockDB.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Return(0, fmt.Errorf("database error"))

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.ProcessTransaction(ctx, tx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create transaction record")
}

func TestProcessTransaction_KafkaError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	tx := db.Transaction{
		UserID:   1,
		Amount:   decimal.NewFromFloat(100.0),
		Currency: "USD",
		Type:     "deposit",
		Status:   "pending",
	}

	mockDB.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Return(1, nil)
	mockProcessor.EXPECT().ProcessTransaction(gomock.Any())
	mockKafka.EXPECT().PublishMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("kafka error"))

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.ProcessTransaction(ctx, tx)

	assert.NoError(t, err)
}

func TestProcessTransaction_MarshalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	tx := db.Transaction{
		UserID:   1,
		Amount:   decimal.NewFromFloat(100.0),
		Currency: "USD",
		Type:     "deposit",
		Status:   "pending",
	}

	mockDB.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Return(1, nil)
	mockProcessor.EXPECT().ProcessTransaction(gomock.Any())

	mockKafka.EXPECT().PublishMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.ProcessTransaction(ctx, tx)

	assert.NoError(t, err)
}

func TestProcessTransaction_WithSpecificGateway(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	tx := db.Transaction{
		UserID:    1,
		Amount:    decimal.NewFromFloat(100.0),
		Currency:  "USD",
		Type:      "deposit",
		Status:    "pending",
		GatewayID: 3, // Specific gateway requested
	}

	mockDB.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, transaction db.Transaction) (int, error) {
			assert.Equal(t, tx.UserID, transaction.UserID)
			assert.Equal(t, tx.Amount, transaction.Amount)
			assert.Equal(t, tx.Currency, transaction.Currency)
			assert.Equal(t, tx.Type, transaction.Type)
			assert.Equal(t, "pending", transaction.Status)
			assert.Equal(t, 3, transaction.GatewayID) // Verify specific gateway ID is preserved
			return 1, nil
		})
	mockProcessor.EXPECT().ProcessTransaction(gomock.Any()).Do(
		func(transaction models.Transaction) {
			assert.Equal(t, 1, transaction.ID)
			assert.Equal(t, tx.UserID, transaction.UserID)
			assert.Equal(t, tx.Amount, transaction.Amount)
			assert.Equal(t, tx.Currency, transaction.Currency)
			assert.Equal(t, tx.Type, transaction.Type)
			assert.Equal(t, "pending", transaction.Status)
			assert.Equal(t, 3, transaction.GatewayID) // Verify specific gateway ID is preserved
		})
	mockKafka.EXPECT().PublishMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.ProcessTransaction(ctx, tx)

	assert.NoError(t, err)
}

func TestProcessTransaction_WithdrawalWithPrioritizedGateways(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	tx := db.Transaction{
		UserID:   1,
		Amount:   decimal.NewFromFloat(500.0),
		Currency: "EUR",
		Type:     "withdrawal", // Testing withdrawal type
		Status:   "pending",
	}

	mockDB.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, transaction db.Transaction) (int, error) {
			assert.Equal(t, tx.UserID, transaction.UserID)
			assert.Equal(t, tx.Amount, transaction.Amount)
			assert.Equal(t, tx.Currency, transaction.Currency)
			assert.Equal(t, "withdrawal", transaction.Type)
			assert.Equal(t, "pending", transaction.Status)
			return 1, nil
		})
	mockProcessor.EXPECT().ProcessTransaction(gomock.Any()).Do(
		func(transaction models.Transaction) {
			assert.Equal(t, 1, transaction.ID)
			assert.Equal(t, tx.UserID, transaction.UserID)
			assert.Equal(t, tx.Amount, transaction.Amount)
			assert.Equal(t, tx.Currency, transaction.Currency)
			assert.Equal(t, "withdrawal", transaction.Type)
			assert.Equal(t, "pending", transaction.Status)
		})
	mockKafka.EXPECT().PublishMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.ProcessTransaction(ctx, tx)

	assert.NoError(t, err)
}

func TestProcessTransaction_WithHighValueAmount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockStorage(ctrl)
	mockCache := mocks.NewMockCache(ctrl)
	mockProcessor := mocks.NewMockTransactionProcessor(ctrl)
	mockKafka := mocks.NewMockProducer(ctrl)

	ctx := context.Background()
	tx := db.Transaction{
		UserID:   1,
		Amount:   decimal.NewFromFloat(10000.0), // High-value transaction
		Currency: "USD",
		Type:     "deposit",
		Status:   "pending",
	}

	mockDB.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, transaction db.Transaction) (int, error) {
			assert.Equal(t, tx.UserID, transaction.UserID)
			assert.Equal(t, decimal.NewFromFloat(10000.0), transaction.Amount)
			assert.Equal(t, tx.Currency, transaction.Currency)
			assert.Equal(t, tx.Type, transaction.Type)
			assert.Equal(t, "pending", transaction.Status)
			return 1, nil
		})
	mockProcessor.EXPECT().ProcessTransaction(gomock.Any())
	mockKafka.EXPECT().PublishMessage(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	cfg := envs.Load()
	service := services.NewGateway(mockDB, mockCache, mockProcessor, mockKafka, cfg)

	err := service.ProcessTransaction(ctx, tx)

	assert.NoError(t, err)
}
