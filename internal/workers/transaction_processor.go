package workers

import (
	"context"
	"sync"

	"payment-gateway/configs/logger"
	"payment-gateway/db"
	"payment-gateway/internal/gateway"
	"payment-gateway/internal/models"
)

type TransactionProcessor interface {
	Start(ctx context.Context)
	Stop()
	ProcessTransaction(tx models.Transaction)
}

var _ TransactionProcessor = (*Processor)(nil)

type Processor struct {
	DB            db.Storage
	WorkerCount   int
	jobs          chan models.Transaction
	gatewayClient gateway.GatewayClient
	wg            sync.WaitGroup
}

func NewTransactionProcessor(
	db db.Storage,
	workerCount int,
	gatewayClient gateway.GatewayClient,
) TransactionProcessor {
	return &Processor{
		DB:            db,
		WorkerCount:   workerCount,
		jobs:          make(chan models.Transaction, 100),
		gatewayClient: gatewayClient,
	}
}

func (p *Processor) Start(ctx context.Context) {
	for i := 0; i < p.WorkerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx)
	}
}

func (p *Processor) Stop() {
	close(p.jobs)
	p.wg.Wait()
}

func (p *Processor) ProcessTransaction(tx models.Transaction) {
	p.jobs <- tx
}

func (p *Processor) worker(ctx context.Context) {
	defer p.wg.Done()

	for tx := range p.jobs {
		user, err := p.DB.GetUserByID(ctx, tx.UserID)
		if err != nil {
			logger.Error("Failed to get user for transaction", "id", tx.ID, "error", err)
			p.markTransactionFailed(ctx, tx.ID, "Failed to get user information")
			continue
		}

		gateways, err := p.DB.GetGatewaysByCountry(ctx, user.CountryID)
		if err != nil {
			logger.Error("Failed to get gateways for transaction", "id", tx.ID, "error", err)
			p.markTransactionFailed(ctx, tx.ID, "Failed to get payment gateways")
			continue
		}

		if len(gateways) == 0 {
			// todo handle properly
			gateways = []db.Gateway{{ID: tx.GatewayID}}
		}

		var lastError error
		var succeeded bool

		for _, gateway := range gateways {
			currentTx := tx
			currentTx.GatewayID = gateway.ID

			gatewayTxnID, err := p.gatewayClient.ProcessPayment(ctx, currentTx)

			if err != nil {
				logger.Warn("Gateway processing failed, trying fallback",
					"txID", tx.ID,
					"gatewayID", gateway.ID,
					"error", err)
				lastError = err
				continue
			}

			// todo wrap to transaction or single execution
			err = p.DB.UpdateTransactionStatus(ctx, tx.ID, "processing", gatewayTxnID, "")
			if err != nil {
				logger.Warn("Failed to update transaction state", "id", tx.ID, "error", err)
			}

			err = p.DB.UpdateTransactionGateway(ctx, tx.ID, gateway.ID)
			if err != nil {
				logger.Warn("Failed to update transaction gateway", "id", tx.ID, "gatewayID", gateway.ID, "error", err)
			}

			succeeded = true

			break
		}

		if !succeeded {
			errorMsg := "All payment gateways failed"
			if lastError != nil {
				errorMsg = lastError.Error()
			}

			p.markTransactionFailed(ctx, tx.ID, errorMsg)
		}
	}
}

func (p *Processor) markTransactionFailed(ctx context.Context, txID int, errorMsg string) {
	err := p.DB.UpdateTransactionStatus(ctx, txID, "failed", "", errorMsg)
	if err != nil {
		logger.Warn("Failed to update transaction status", "id", txID, "error", err)
	}
}
