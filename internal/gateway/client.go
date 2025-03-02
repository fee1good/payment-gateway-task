package gateway

import (
	"context"
	"strconv"

	"payment-gateway/configs/envs"
	"payment-gateway/configs/logger"
	"payment-gateway/internal/models"
	"payment-gateway/internal/utils"
)

type GatewayClient interface {
	ProcessPayment(ctx context.Context, tx models.Transaction) (string, error)
}

var _ GatewayClient = (*Client)(nil)

type Client struct {
	maxRetries int
}

func Stripe(config *envs.Config) GatewayClient {
	return &Client{
		maxRetries: config.Retry.MaxRetries,
	}
}

// ProcessPayment todo use context when real implementation is added
func (c *Client) ProcessPayment(_ context.Context, tx models.Transaction) (string, error) {
	var gatewayTxnID string
	var processErr error

	err := utils.RetryOperation(func() error {
		// This simulates the actual call to the payment gateway API
		gatewayTxnID = "gateway-txn-" + strconv.Itoa(tx.ID)

		logger.Info("Attempting to process payment with gateway",
			"txID", tx.ID,
			"gatewayID", tx.GatewayID,
			"attempt", gatewayTxnID)

		// In a real implementation, this would make an HTTP request to the gateway
		// and could return an error if the request fails

		// For demonstration, we'll always succeed
		processErr = nil
		return processErr
	}, c.maxRetries)

	if err != nil {
		logger.Error("All attempts to process payment failed",
			"txID", tx.ID,
			"gatewayID", tx.GatewayID,
			"error", err)
		return "", err
	}

	logger.Info("Successfully processed payment with gateway",
		"txID", tx.ID,
		"gatewayID", tx.GatewayID,
		"gatewayTxnID", gatewayTxnID)

	return gatewayTxnID, nil
}
