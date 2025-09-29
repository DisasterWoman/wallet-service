package service

import (
	"context"
	"github.com/DisasterWoman/wallet-service/internal/models"
	"github.com/google/uuid"
)

type WalletService interface {
	UpdateBalance(ctx context.Context, req *models.OperationRequest) error
	GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error)
}