package models

import (
	"errors"
	"github.com/google/uuid"
)

var (
	ErrInvalidAmount     = errors.New("amount must be positive")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

type OperationType string

const (
	Deposit  OperationType = "DEPOSIT"
	Withdraw OperationType = "WITHDRAW"
)

type Wallet struct {
	ID      uuid.UUID `json:"walletId" db:"id"`
	Balance int64     `json:"balance" db:"balance"`
}

type OperationRequest struct {
	WalletID     uuid.UUID     `json:"walletId"`
	OperationType OperationType `json:"operationType"`
	Amount       int64         `json:"amount"`
}

func (r *OperationRequest) Validate() error {
	if r.Amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}
