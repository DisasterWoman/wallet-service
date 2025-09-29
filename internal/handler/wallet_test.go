package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DisasterWoman/wallet-service/internal/models"
	"github.com/DisasterWoman/wallet-service/internal/repository"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) UpdateBalance(ctx context.Context, req *models.OperationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockService) GetBalance(ctx context.Context, walletID uuid.UUID) (int64, error) {
	args := m.Called(ctx, walletID)
	return args.Get(0).(int64), args.Error(1)
}

func TestWalletHandler_UpdateWalletBalance_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewWalletHandler(mockService)

	walletID := uuid.New()
	reqBody := models.OperationRequest{
		WalletID:      walletID,
		OperationType: models.Deposit,
		Amount:        1000,
	}

	mockService.On("UpdateBalance", mock.Anything, &reqBody).Return(nil)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/wallet", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.UpdateWalletBalance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, "success", response["status"])
	
	mockService.AssertExpectations(t)
}

func TestWalletHandler_UpdateWalletBalance_InvalidJSON(t *testing.T) {
	mockService := new(MockService)
	handler := NewWalletHandler(mockService)

	// Невалидный JSON
	body := []byte(`{"invalid": json`)
	req := httptest.NewRequest("POST", "/api/v1/wallet", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.UpdateWalletBalance(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "UpdateBalance")
}

func TestWalletHandler_UpdateWalletBalance_InsufficientFunds(t *testing.T) {
	mockService := new(MockService)
	handler := NewWalletHandler(mockService)

	walletID := uuid.New()
	reqBody := models.OperationRequest{
		WalletID:      walletID,
		OperationType: models.Withdraw,
		Amount:        1000,
	}

	mockService.On("UpdateBalance", mock.Anything, &reqBody).Return(models.ErrInsufficientFunds)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/wallet", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.UpdateWalletBalance(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	mockService.AssertExpectations(t)
}

func TestWalletHandler_UpdateWalletBalance_WalletNotFound(t *testing.T) {
	mockService := new(MockService)
	handler := NewWalletHandler(mockService)

	walletID := uuid.New()
	reqBody := models.OperationRequest{
		WalletID:      walletID,
		OperationType: models.Deposit,
		Amount:        1000,
	}

	mockService.On("UpdateBalance", mock.Anything, &reqBody).Return(repository.ErrWalletNotFound)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/wallet", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.UpdateWalletBalance(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	mockService.AssertExpectations(t)
}

func TestWalletHandler_UpdateWalletBalance_InvalidAmount(t *testing.T) {
	mockService := new(MockService)
	handler := NewWalletHandler(mockService)

	walletID := uuid.New()
	reqBody := models.OperationRequest{
		WalletID:      walletID,
		OperationType: models.Deposit,
		Amount:        -100, // Невалидная сумма
	}

	mockService.On("UpdateBalance", mock.Anything, &reqBody).Return(models.ErrInvalidAmount)

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/wallet", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.UpdateWalletBalance(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertExpectations(t)
}

func TestWalletHandler_GetWalletBalance_Success(t *testing.T) {
	mockService := new(MockService)
	handler := NewWalletHandler(mockService)

	walletID := uuid.New()
	expectedBalance := int64(1500)

	mockService.On("GetBalance", mock.Anything, walletID).Return(expectedBalance, nil)

	req := httptest.NewRequest("GET", "/api/v1/wallets/"+walletID.String(), nil)
	rr := httptest.NewRecorder()

	// Используем mux для извлечения параметров
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallets/{walletId}", handler.GetWalletBalance)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]int64
	json.Unmarshal(rr.Body.Bytes(), &response)
	assert.Equal(t, expectedBalance, response["balance"])
	
	mockService.AssertExpectations(t)
}

func TestWalletHandler_GetWalletBalance_WalletNotFound(t *testing.T) {
	mockService := new(MockService)
	handler := NewWalletHandler(mockService)

	walletID := uuid.New()

	mockService.On("GetBalance", mock.Anything, walletID).Return(int64(0), repository.ErrWalletNotFound)

	req := httptest.NewRequest("GET", "/api/v1/wallets/"+walletID.String(), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallets/{walletId}", handler.GetWalletBalance)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	mockService.AssertExpectations(t)
}

func TestWalletHandler_GetWalletBalance_InvalidUUID(t *testing.T) {
	mockService := new(MockService)
	handler := NewWalletHandler(mockService)

	// Невалидный UUID
	req := httptest.NewRequest("GET", "/api/v1/wallets/invalid-uuid", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallets/{walletId}", handler.GetWalletBalance)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mockService.AssertNotCalled(t, "GetBalance")
}

func TestWalletHandler_GetWalletBalance_InternalError(t *testing.T) {
	mockService := new(MockService)
	handler := NewWalletHandler(mockService)

	walletID := uuid.New()

	mockService.On("GetBalance", mock.Anything, walletID).Return(int64(0), assert.AnError)

	req := httptest.NewRequest("GET", "/api/v1/wallets/"+walletID.String(), nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/wallets/{walletId}", handler.GetWalletBalance)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	mockService.AssertExpectations(t)
}