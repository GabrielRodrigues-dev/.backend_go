package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/GabrielRodrigues-dev/backend_go/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WalletRepository isola a comunicação com o banco de dados.

type WalletRepository struct {
	db *pgxpool.Pool
}

// NewWalletRepository é o construtor que o Uber Fx vai usar.
func NewWalletRepository(db *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{
		db: db,
	}
}

// Create insere uma nova carteira no banco de dados.
func (r *WalletRepository) Create(ctx context.Context, w domain.Wallet) error {

	query := `
		INSERT INTO wallets (id, player_id, balance_cents, currency, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	// db.Exec executa a query sem esperar linhas de retorno.
	_, err := r.db.Exec(ctx, query,
		w.ID(),
		w.PlayerID(),
		w.Balance().Cents(), // Pega os centavos seguros (int64)!
		w.Balance().Currency(),
		w.Version(),
		w.CreatedAt(),
		w.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("erro ao inserir carteira no banco: %w", err)
	}

	return nil
}

// FindByID busca uma carteira no banco de dados e a reconstrói.
func (r *WalletRepository) FindByID(ctx context.Context, id string) (domain.Wallet, error) {
	var playerId, currency string
	var balanceCents int64
	var version int
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(ctx, `SELECT player_id, balance_cents, currency, version, created_at, updated_at FROM wallets WHERE id = $1`, id).
		Scan(&playerId, &balanceCents, &currency, &version, &createdAt, &updatedAt)

	if err != nil {
		return domain.Wallet{}, fmt.Errorf("carteira nao encontrada: %w", err)
	}

	return domain.RestoreWallet(id, playerId, balanceCents, currency, version, createdAt, updatedAt), nil
}

// ProcessTransaction efetua o débito/crédito na carteira e regista o ledger numa única transação atómica.
func (r *WalletRepository) ProcessTransaction(ctx context.Context, walletID string, ledgerEntry domain.LedgerEntry, newVersion int) error {
	// 1. Inicia a Transação (Se algo falhar no meio, nada é guardado)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transacao: %w", err)
	}
	// O defer garante que, se houver um erro, a transação faz ROLLBACK automaticamente.
	defer tx.Rollback(ctx)

	// 2. Lock Pessimista
	// Tranca a linha desta carteira específica. Qualquer outra aposta que chegar agora ficará "à espera" até esta transação terminar.
	var currentVersion int
	err = tx.QueryRow(ctx, `SELECT version FROM wallets WHERE id = $1 FOR NO KEY UPDATE`, walletID).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("erro ao bloquear carteira ou carteira nao encontrada: %w", err)
	}

	// 3. Atualiza o saldo da carteira COM CONTROLE OTIMISTA
	tag, err := tx.Exec(ctx, `
		UPDATE wallets 
		SET balance_cents = $1, version = $2, updated_at = $3
		WHERE id = $4 AND version = $5
	`, ledgerEntry.BalanceAfter().Cents(), newVersion, ledgerEntry.CreatedAt(), walletID, newVersion-1)

	if err != nil {
		return fmt.Errorf("erro ao atualizar carteira: %w", err)
	}

	// Se nenhuma linha foi afetada, significa que outra transação já atualizou a versão!
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("conflito de concorrência: a carteira foi atualizada por outra transacao simultanea")
	}

	// 4. Cria o registo no Ledger
	_, err = tx.Exec(ctx, `
		INSERT INTO wallet_ledger_entries 
		(id, wallet_id, transaction_id, direction, amount_cents, balance_before_cents, balance_after_cents, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		ledgerEntry.ID(),
		ledgerEntry.WalletID(),
		ledgerEntry.TransactionID(),
		ledgerEntry.Direction(),
		ledgerEntry.Amount().Cents(),
		ledgerEntry.BalanceBefore().Cents(),
		ledgerEntry.BalanceAfter().Cents(),
		ledgerEntry.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("erro ao inserir ledger: %w", err)
	}

	// 5. Confirma a Transação (Guarda de forma permanente)
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("erro no commit da transacao: %w", err)
	}

	return nil
}
