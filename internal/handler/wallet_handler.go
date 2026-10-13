package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/GabrielRodrigues-dev/backend_go/internal/domain"
	"github.com/GabrielRodrigues-dev/backend_go/internal/repository"
)

// Estrutura que representa o JSON.
type BetRequest struct {
	WalletID      string `json:"wallet_id"`
	TransactionID string `json:"transaction_id"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
}

type WalletHandler struct {
	repo *repository.WalletRepository
}

// Construtor para o Uber Fx
func NewWalletHandler(repo *repository.WalletRepository) *WalletHandler {
	return &WalletHandler{repo: repo}
}

// Rota que processa uma aposta HTTP
func (h *WalletHandler) PlaceBet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 1. Ler o JSON do corpo da requisição
	var req BetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 2. Convertemos as strings para o Value Objects seguros
	betAmount, err := domain.NewMoneyFromString(req.Amount, req.Currency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Busca a carteira no banco
	ctx := r.Context()
	wallet, err := h.repo.FindByID(ctx, req.WalletID)
	if err != nil {
		http.Error(w, "Wallet not found", http.StatusNotFound)
		return
	}

	// 4. Regras de negócio (Debitar)
	initialMoney := wallet.Balance()
	err = wallet.Debit(betAmount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity) // 422 para erros de regra de negócio
		return
	}

	// 5. Cria o Registo no Ledger e salva no banco atômicamente
	ledgerEntry := domain.NewLedgerEntry(
		req.TransactionID, wallet.ID(), "game-provider-req", "DEBIT", betAmount, initialMoney, wallet.Balance(),
	)

	err = h.repo.ProcessTransaction(ctx, wallet.ID(), ledgerEntry, wallet.Version())
	if err != nil {
		// Se der erro aqui, é provável que seja o nosso conflito de concorrência!
		http.Error(w, "Conflict processing transaction", http.StatusConflict)
		return
	}

	// 6. Devolvemos 200 OK e o novo saldo.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "success",
		"new_balance": wallet.Balance().Amount(),
		"version":     fmt.Sprintf("%d", wallet.Version()),
	})
}
