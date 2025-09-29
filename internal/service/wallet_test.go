package service

import (
	"context"
	"testing"

	"github.com/DisasterWoman/wallet-service/internal/models"
	"github.com/DisasterWoman/wallet-service/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	args := m.Called(ctx, walletID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) UpdateBalance(ctx context.Context, walletID uuid.UUID, amount int64) error {
	args := m.Called(ctx, walletID, amount)
	return args.Error(0)
}

func TestWalletService_UpdateBalance_Deposit(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewWalletService(mockRepo)

	walletID := uuid.New()
	req := &models.OperationRequest{
		WalletID:      walletID,
		OperationType: models.Deposit,
		Amount:        1000,
	}

	mockRepo.On("UpdateBalance", mock.Anything, walletID, int64(1000)).Return(nil)

	err := service.UpdateBalance(context.Background(), req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWalletService_UpdateBalance_Withdraw(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewWalletService(mockRepo)

	walletID := uuid.New()
	req := &models.OperationRequest{
		WalletID:      walletID,
		OperationType: models.Withdraw,
		Amount:        500,
	}

	mockRepo.On("UpdateBalance", mock.Anything, walletID, int64(-500)).Return(nil)

	err := service.UpdateBalance(context.Background(), req)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestWalletService_UpdateBalance_InvalidAmount(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewWalletService(mockRepo)

	walletID := uuid.New()
	req := &models.OperationRequest{
		WalletID:      walletID,
		OperationType: models.Deposit,
		Amount:        -100, // Невалидная сумма
	}

	err := service.UpdateBalance(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, models.ErrInvalidAmount, err)
	mockRepo.AssertNotCalled(t, "UpdateBalance")
}

func TestWalletService_GetBalance(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewWalletService(mockRepo)

	walletID := uuid.New()
	expectedBalance := int64(1500)

	mockRepo.On("GetBalance", mock.Anything, walletID).Return(expectedBalance, nil)

	balance, err := service.GetBalance(context.Background(), walletID)

	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)
	mockRepo.AssertExpectations(t)
}

func TestWalletService_GetBalance_Error(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewWalletService(mockRepo)

	walletID := uuid.New()

	mockRepo.On("GetBalance", mock.Anything, walletID).Return(int64(0), repository.ErrWalletNotFound)

	balance, err := service.GetBalance(context.Background(), walletID)

	assert.Error(t, err)
	assert.Equal(t, repository.ErrWalletNotFound, err)
	assert.Equal(t, int64(0), balance)
	mockRepo.AssertExpectations(t)
}