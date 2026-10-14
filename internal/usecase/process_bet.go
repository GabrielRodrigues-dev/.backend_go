package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/GabrielRodrigues-dev/backend_go/internal/domain"
	"github.com/GabrielRodrigues-dev/backend_go/internal/repository"
)

type ProcessBetUseCase struct {
	repo *repository.WalletRepository
}

func NewProcessBetUseCase(repo *repository.WalletRepository) *ProcessBetUseCase {
	return &ProcessBetUseCase{repo: repo}
}

// InputData representa o que precisamos para processar uma aposta
type ProcessBetInput struct {
	WalletID       string
	TransactionID  string
	Amount         string
	Currency       string
	IdempotencyKey string
}

func (uc *ProcessBetUseCase) Execute(ctx context.Context, input ProcessBetInput) error {
	betAmount, err := domain.NewMoneyFromString(input.Amount, input.Currency)
	if err != nil {
		return err
	}

	wallet, err := uc.repo.FindByID(ctx, input.WalletID)
	if err != nil {
		return err
	}

	initialMoney := wallet.Balance()
	err = wallet.Debit(betAmount)
	if err != nil {
		return err
	}

	ledgerEntry := domain.NewLedgerEntry(
		"ledger-"+input.TransactionID, wallet.ID(), input.TransactionID, "DEBIT", betAmount, initialMoney, wallet.Balance(),
	)

	// Cria o Evento para a Outbox
	payload := domain.WalletBalanceChangedPayload{
		WalletID:      wallet.ID(),
		TransactionID: input.TransactionID,
		Direction:     "DEBIT",
		Money:         betAmount.Amount(),
		BalanceBefore: initialMoney.Amount(),
		BalanceAfter:  wallet.Balance().Amount(),
		WalletVersion: wallet.Version(),
	}

	event, err := domain.NewOutboxEvent(
		fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		wallet.ID(),
		"WalletBalanceChanged",
		payload,
	)
	if err != nil {
		return err
	}

	// Executa a transação massiva
	err = uc.repo.ProcessTransaction(ctx, wallet.ID(), ledgerEntry, wallet.Version(), input.IdempotencyKey, []domain.OutboxEvent{event})
	if err != nil {
		return err
	}

	return nil
}
