package repository

import (
	"context"
	"github.com/google/uuid"
)

type Repository interface {
	GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error)
	UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64) error
}