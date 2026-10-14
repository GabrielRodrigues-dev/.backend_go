package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GabrielRodrigues-dev/backend_go/internal/domain"
	"github.com/GabrielRodrigues-dev/backend_go/internal/repository"
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
	repo       *repository.WalletRepository
}

func NewWalletHandler(processBet *usecase.ProcessBetUseCase, repo *repository.WalletRepository) *WalletHandler {
	return &WalletHandler{processBet: processBet, repo: repo}
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

	// A Chave de Idempotência que vem do cabeçalho HTTP.
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

// Estrutura exigida pelo desafio para criação de carteira.
type CreateWalletRequest struct {
	PlayerID       string `json:"playerId"`
	InitialBalance struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"initialBalance"`
}

func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Converte o saldo inicial
	initialMoney, err := domain.NewMoneyFromString(req.InitialBalance.Amount, req.InitialBalance.Currency)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Gera um ID único para a carteira e cria a entidade[cite: 12]
	walletID := fmt.Sprintf("wallet-%d", time.Now().UnixNano())
	wallet := domain.NewWallet(walletID, req.PlayerID, initialMoney)

	var ledger *domain.LedgerEntry
	var events []domain.OutboxEvent

	// Se o saldo for maior que zero, exige criação do ledger "OPENING" e eventos[cite: 12]
	if initialMoney.Cents() > 0 {
		// Encurtando o ID para caber no VARCHAR(36)
		nowMs := time.Now().UnixMilli()
		openingTxID := fmt.Sprintf("txn-op-%d", nowMs)

		l := domain.NewLedgerEntry(
			fmt.Sprintf("ldg-%d", nowMs), wallet.ID(), openingTxID, "CREDIT",
			initialMoney, domain.ZeroMoney(initialMoney.Currency()), initialMoney,
		)
		ledger = &l

		payload := domain.WalletBalanceChangedPayload{
			WalletID:      wallet.ID(),
			TransactionID: openingTxID,
			Direction:     "CREDIT",
			Money:         initialMoney.Amount(),
			BalanceBefore: "0.00",
			BalanceAfter:  initialMoney.Amount(),
			WalletVersion: wallet.Version(),
		}

		event, _ := domain.NewOutboxEvent(fmt.Sprintf("evt-%d", nowMs), wallet.ID(), "WalletBalanceChanged", payload)
		events = append(events, event)
	}

	// Executa a criação no repositório acedendo ao cofre através do UseCase (ou injetando o repo diretamente)
	err = h.repo.CreateWalletTx(r.Context(), wallet, ledger, events)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id":       wallet.ID(),
		"playerId": wallet.PlayerID(),
		"balance":  map[string]string{"amount": wallet.Balance().Amount(), "currency": wallet.Balance().Currency()},
		"version":  wallet.Version(),
	})
}
