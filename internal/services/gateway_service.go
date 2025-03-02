package services

import (
	"context"
	"encoding/json"
	"fmt"

	"payment-gateway/configs/envs"
	"payment-gateway/configs/logger"
	"payment-gateway/db"
	"payment-gateway/internal/cache"
	"payment-gateway/internal/kafka"
	"payment-gateway/internal/models"
	"payment-gateway/internal/utils"
	"payment-gateway/internal/workers"
)

type GatewayServiceInterface interface {
	ProcessTransaction(ctx context.Context, tx db.Transaction) error
	HandleCallback(ctx context.Context, gatewayTxnID string, status string, transactionID int) error
	GetTransactionStatus(ctx context.Context, txID int) (db.Transaction, error)
}

var _ GatewayServiceInterface = (*GatewayService)(nil)

type GatewayService struct {
	DB                   db.Storage
	Cache                cache.Cache
	TransactionProcessor workers.TransactionProcessor
	kafkaProducer        kafka.Producer
	cfg                  *envs.Config
}

func NewGateway(
	db db.Storage,
	cache cache.Cache,
	processor workers.TransactionProcessor,
	kafkaProducer kafka.Producer,
	cfg *envs.Config,
) GatewayServiceInterface {
	return &GatewayService{
		DB:                   db,
		Cache:                cache,
		TransactionProcessor: processor,
		kafkaProducer:        kafkaProducer,
		cfg:                  cfg,
	}
}

func (s *GatewayService) ProcessTransaction(ctx context.Context, tx db.Transaction) error {
	txID, err := s.DB.CreateTransaction(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to create transaction record: %v", err)
	}
	tx.ID = txID

	modelsTx := models.Transaction{
		ID:        txID,
		UserID:    tx.UserID,
		Amount:    tx.Amount,
		Currency:  tx.Currency,
		Type:      tx.Type,
		Status:    "pending",
		GatewayID: tx.GatewayID,
	}

	s.TransactionProcessor.ProcessTransaction(modelsTx)

	txData := models.Transaction{
		ID:        txID,
		UserID:    tx.UserID,
		Amount:    tx.Amount,
		Currency:  tx.Currency,
		Type:      tx.Type,
		Status:    "pending",
		GatewayID: tx.GatewayID,
	}

	txDataBytes, err := json.Marshal(txData)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction data: %v", err)
	}

	maskedData := utils.MaskData(txDataBytes)

	err = utils.ExecuteWithCircuitBreaker(func() error {
		if s.kafkaProducer != nil {
			return s.kafkaProducer.PublishMessage(ctx, s.cfg.Kafka.TransactionsTopic, []byte(maskedData))
		}
		logger.Warn("Kafka producer not initialized, skipping message publication")
		return nil
	})

	if err != nil {
		// Continue processing even if Kafka publish fails
		// todo add logs
	}

	return nil
}

func (s *GatewayService) HandleCallback(ctx context.Context, gatewayTxnID string, status string, transactionID int) error {
	// todo this should be wrapped in a transaction
	tx, err := s.DB.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("transaction not found: %v", err)
	}

	// todo move to enum
	if tx.Status == "completed" || tx.Status == "failed" {
		return fmt.Errorf("transaction already in final state: %s", tx.Status)
	}

	internalStatus := mapGatewayStatus(status)

	err = s.DB.UpdateTransactionStatus(ctx, transactionID, internalStatus, gatewayTxnID, "")
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %v", err)
	}

	return nil
}

func mapGatewayStatus(gatewayStatus string) string {
	switch gatewayStatus {
	case "success", "completed", "approved":
		return "completed"
	case "failed", "declined", "rejected":
		return "failed"
	default:
		return "pending"
	}
}

func (s *GatewayService) GetTransactionStatus(ctx context.Context, txID int) (db.Transaction, error) {
	return s.DB.GetTransactionByID(ctx, txID)
}
