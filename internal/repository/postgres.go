package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/DisasterWoman/wallet-service/internal/models"
	
)

var (
	ErrWalletNotFound = errors.New("wallet not found")
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	var balance int64
	err := r.db.QueryRowContext(
		ctx,
		"SELECT balance FROM wallets WHERE id = $1",
		walletID,
	).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrWalletNotFound
	}
	return balance, err
}

func (r *PostgresRepository) UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() 

	var currentBalance int64
	err = tx.QueryRowContext(
		ctx,
		"SELECT balance FROM wallets WHERE id = $1 FOR UPDATE", 
		walletID,
	).Scan(&currentBalance)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrWalletNotFound
	}
	if err != nil {
		return err
	}

	if amount < 0 && currentBalance+amount < 0 {
		return models.ErrInsufficientFunds
	}

	_, err = tx.ExecContext(
		ctx,
		"UPDATE wallets SET balance = balance + $1 WHERE id = $2",
		amount,
		walletID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
