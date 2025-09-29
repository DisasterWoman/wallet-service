package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/DisasterWoman/wallet-service/internal/models"
	"github.com/DisasterWoman/wallet-service/internal/repository"
)

type WalletService struct {
	repo *repository.PostgresRepository
}

func NewWalletService(repo *repository.PostgresRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) UpdateBalance(ctx context.Context, req *models.OperationRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	amount := req.Amount
	if req.OperationType == models.Withdraw {
		amount = -amount
	}

	return s.repo.UpdateBalance(ctx, req.WalletID, amount)
}

func (s *WalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	return s.repo.GetBalance(ctx, walletID)
}
