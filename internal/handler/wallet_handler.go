package handler

import (
	"encoding/json"
	"net/http"

	"github.com/GabrielRodrigues-dev/backend_go/internal/usecase"
)

// Estrutura que representa o JSON.
type BetRequest struct {
	WalletID      string `json:"wallet_id"`
	TransactionID string `json:"transaction_id"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
}

type WalletHandler struct {
	processBet *usecase.ProcessBetUseCase
}

func NewWalletHandler(processBet *usecase.ProcessBetUseCase) *WalletHandler {
	return &WalletHandler{processBet: processBet}
}

func (h *WalletHandler) PlaceBet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// A Chave de Idempotência que vem do cabeçalho HTTP[cite: 11]
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		http.Error(w, "Missing Idempotency-Key header", http.StatusBadRequest)
		return
	}

	input := usecase.ProcessBetInput{
		WalletID:       req.WalletID,
		TransactionID:  req.TransactionID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		IdempotencyKey: idempotencyKey,
	}

	err := h.processBet.Execute(r.Context(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
	})
}
