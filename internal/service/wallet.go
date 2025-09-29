package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/DisasterWoman/wallet-service/internal/models"
	"github.com/DisasterWoman/wallet-service/internal/repository"
)

type walletService struct {
	repo repository.Repository
}

func NewWalletService(repo repository.Repository) WalletService {  
	return &walletService{repo: repo}
}

func (s *walletService) UpdateBalance(ctx context.Context, req *models.OperationRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	amount := req.Amount
	if req.OperationType == models.Withdraw {
		amount = -amount
	}

	return s.repo.UpdateBalance(ctx, req.WalletID, amount)
}

func (s *walletService) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	return s.repo.GetBalance(ctx, walletID)
}