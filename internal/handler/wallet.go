package handler

import (
	"encoding/json"
	"net/http"

	"github.com/DisasterWoman/wallet-service/internal/models"
	"github.com/DisasterWoman/wallet-service/internal/repository"
	"github.com/DisasterWoman/wallet-service/internal/service"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type WalletHandler struct {
	service service.WalletService
}

func NewWalletHandler(service service.WalletService) *WalletHandler {
	return &WalletHandler{service: service}
}

func (h *WalletHandler) UpdateWalletBalance(w http.ResponseWriter, r *http.Request) {
	var req models.OperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateBalance(r.Context(), &req); err != nil {
		switch err {
		case models.ErrInvalidAmount:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case models.ErrInsufficientFunds, repository.ErrWalletNotFound:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *WalletHandler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletID, err := uuid.Parse(vars["walletId"])
	if err != nil {
		http.Error(w, "invalid wallet ID", http.StatusBadRequest)
		return
	}

	balance, err := h.service.GetBalance(r.Context(), walletID)
	if err != nil {
		if err == repository.ErrWalletNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(map[string]int64{"balance": balance})
}