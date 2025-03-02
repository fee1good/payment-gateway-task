package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
	"payment-gateway/configs/logger"
)

var db *sql.DB

type User struct {
	ID        int
	Username  string
	Email     string
	CountryID int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Gateway struct {
	ID                  int
	Name                string
	DataFormatSupported string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type Country struct {
	ID        int
	Name      string
	Code      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	ID           int
	UserID       int
	Amount       decimal.Decimal
	Currency     string
	Type         string // "deposit" or "withdrawal"
	Status       string // "pending", "completed", "failed"
	GatewayID    int
	GatewayTxnID string
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
}

type Storage interface {
	GetUserByID(ctx context.Context, id int) (User, error)
	UpdateTransactionStatus(ctx context.Context, txID int, status, gatewayTxnID, errorMsg string) error
	CreateTransaction(ctx context.Context, tx Transaction) (int, error)
	GetTransactionByID(ctx context.Context, id int) (Transaction, error)
	GetGatewaysByCountry(ctx context.Context, countryID int) ([]Gateway, error)
	UpdateTransactionGateway(ctx context.Context, txID int, gatewayID int) error
}

type Postgres struct {
	db *sql.DB
}

func NewDBHandler(db *sql.DB) Storage {
	return &Postgres{db: db}
}

func InitDB(dataSourceName string) (*sql.DB, error) {
	var err error

	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	logger.Info("Successfully connected to the database.")
	return db, nil
}

func (p *Postgres) CreateTransaction(ctx context.Context, tx Transaction) (int, error) {
	query := `
		INSERT INTO transactions 
		(user_id, amount, currency, type, status, gateway_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id
	`

	var id int
	err := p.db.QueryRowContext(
		ctx,
		query,
		tx.UserID,
		tx.Amount,
		tx.Currency,
		tx.Type,
		tx.Status,
		tx.GatewayID,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create transaction: %v", err)
	}

	return id, nil
}

func (p *Postgres) UpdateTransactionStatus(ctx context.Context, id int, status string, gatewayTxnID string, errorMsg string) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	var existingStatus string
	lockQuery := `SELECT status FROM transactions WHERE id = $1 FOR UPDATE`
	err = tx.QueryRowContext(ctx, lockQuery, id).Scan(&existingStatus)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to lock transaction row: %v", err)
	}

	if existingStatus == "completed" || existingStatus == "failed" {
		tx.Rollback()
		return fmt.Errorf("cannot update transaction in final state: %s", existingStatus)
	}

	query := `
		UPDATE transactions 
		SET status = $1, gateway_txn_id = $2, error_message = $3, updated_at = $4
	`

	args := []interface{}{status, gatewayTxnID, errorMsg, time.Now()}

	if status == "completed" {
		query += `, completed_at = $5 WHERE id = $6`
		now := time.Now()
		args = append(args, now, id)
	} else {
		query += ` WHERE id = $5`
		args = append(args, id)
	}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update transaction status: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (p *Postgres) GetTransactionByID(ctx context.Context, id int) (Transaction, error) {
	query := `
		SELECT id, user_id, amount, currency, type, status, gateway_id, 
		       gateway_txn_id, error_message, created_at, updated_at, completed_at 
		FROM transactions 
		WHERE id = $1
	`

	var tx Transaction
	var gatewayTxnID, errorMsg sql.NullString
	var gatewayID sql.NullInt64
	var completedAt sql.NullTime

	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&tx.ID, &tx.UserID, &tx.Amount, &tx.Currency, &tx.Type, &tx.Status,
		&gatewayID, &gatewayTxnID, &errorMsg, &tx.CreatedAt, &tx.UpdatedAt, &completedAt,
	)

	if err != nil {
		return Transaction{}, fmt.Errorf("failed to get transaction: %v", err)
	}

	if gatewayID.Valid {
		tx.GatewayID = int(gatewayID.Int64)
	}

	if gatewayTxnID.Valid {
		tx.GatewayTxnID = gatewayTxnID.String
	}

	if errorMsg.Valid {
		tx.ErrorMessage = errorMsg.String
	}

	if completedAt.Valid {
		tx.CompletedAt = &completedAt.Time
	}

	return tx, nil
}

func (p *Postgres) GetGatewaysByCountry(ctx context.Context, countryID int) ([]Gateway, error) {
	query := `
		SELECT g.id, g.name, g.data_format_supported, g.created_at, g.updated_at, gc.priority
		FROM gateways g
		JOIN gateway_countries gc ON g.id = gc.gateway_id
		WHERE gc.country_id = $1
		ORDER BY gc.priority ASC
	`

	rows, err := p.db.QueryContext(ctx, query, countryID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gateways: %v", err)
	}
	defer rows.Close()

	var gateways []Gateway
	for rows.Next() {
		var gateway Gateway
		var priority int
		if err := rows.Scan(
			&gateway.ID, &gateway.Name, &gateway.DataFormatSupported,
			&gateway.CreatedAt, &gateway.UpdatedAt, &priority,
		); err != nil {
			return nil, fmt.Errorf("failed to scan gateway: %v", err)
		}
		gateways = append(gateways, gateway)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return gateways, nil
}

func (p *Postgres) GetUserByID(ctx context.Context, id int) (User, error) {
	query := `SELECT id, username, email, country_id, created_at, updated_at FROM users WHERE id = $1`

	var user User
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.CountryID, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, fmt.Errorf("user not found: %v", err)
		}
		return User{}, fmt.Errorf("failed to get user: %v", err)
	}

	return user, nil
}

func (p *Postgres) UpdateTransactionGateway(ctx context.Context, txID int, gatewayID int) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	lockQuery := `SELECT id FROM transactions WHERE id = $1 FOR UPDATE`
	row := tx.QueryRowContext(ctx, lockQuery, txID)
	var id int
	if err = row.Scan(&id); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to lock transaction row: %v", err)
	}

	query := `
		UPDATE transactions 
		SET gateway_id = $1, updated_at = $2
		WHERE id = $3
	`

	_, err = tx.ExecContext(ctx, query, gatewayID, time.Now(), txID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update transaction gateway: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
