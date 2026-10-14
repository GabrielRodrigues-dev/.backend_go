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

// ProcessTransaction efetua o débito/crédito, ledger, inbox e outbox numa única transação atômica.
func (r *WalletRepository) ProcessTransaction(
	ctx context.Context,
	walletID string,
	ledgerEntry domain.LedgerEntry,
	newVersion int,
	idempotencyKey string,
	events []domain.OutboxEvent,
) error {
	// 1. Inicia a Transação (Se algo falhar no meio, nada é guardado)
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transacao: %w", err)
	}
	// O defer garante que, se houver um erro, a transação faz ROLLBACK automaticamente.
	defer tx.Rollback(ctx)

	// INBOX: Previne processamento duplicado da mesma mensagem.
	// O 'ON CONFLICT DO NOTHING' tenta inserir; se o ID já existir, ele não faz nada e retorna 0 linhas afetadas.
	if idempotencyKey != "" {
		tag, err := tx.Exec(ctx, `
			INSERT INTO inbox (message_id, consumer_name, received_at) 
			VALUES ($1, $2, $3) ON CONFLICT DO NOTHING
		`, idempotencyKey, "betting-engine", ledgerEntry.CreatedAt())

		if err != nil {
			return fmt.Errorf("erro ao verificar inbox: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return fmt.Errorf("conflito de idempotencia: chave %s ja processada", idempotencyKey)
		}
	}

	// 2. Lock Pessimista
	// Tranca a linha desta carteira específica. Qualquer outra aposta que chegar agora ficará "à espera" até esta transação terminar.
	var currentVersion int
	err = tx.QueryRow(ctx, `SELECT version FROM wallets WHERE id = $1 FOR NO KEY UPDATE`, walletID).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("erro ao bloquear carteira: %w", err)
	}

	// 3. Atualiza o saldo da carteira COM CONTROLE OTIMISTA
	tag, err := tx.Exec(ctx, `
		UPDATE wallets SET balance_cents = $1, version = $2, updated_at = $3 WHERE id = $4 AND version = $5
	`, ledgerEntry.BalanceAfter().Cents(), newVersion, ledgerEntry.CreatedAt(), walletID, newVersion-1)

	if err != nil || tag.RowsAffected() == 0 {
		return fmt.Errorf("conflito de concorrência: versao invalida ou carteira atualizada")
	}

	// 4. Cria o Registro no Ledger
	_, err = tx.Exec(ctx, `
		INSERT INTO wallet_ledger_entries 
		(id, wallet_id, transaction_id, direction, amount_cents, balance_before_cents, balance_after_cents, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		ledgerEntry.ID(), ledgerEntry.WalletID(), ledgerEntry.TransactionID(), ledgerEntry.Direction(),
		ledgerEntry.Amount().Cents(), ledgerEntry.BalanceBefore().Cents(), ledgerEntry.BalanceAfter().Cents(), ledgerEntry.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("erro ao inserir ledger: %w", err)
	}

	// 5. OUTBOX: Registra os eventos imutáveis[cite: 11].
	for _, ev := range events {
		_, err = tx.Exec(ctx, `
			INSERT INTO outbox (event_id, aggregate_id, event_type, payload, created_at) 
			VALUES ($1, $2, $3, $4, $5)
		`, ev.EventID, ev.AggregateID, ev.EventType, ev.Payload, ev.CreatedAt)
		if err != nil {
			return fmt.Errorf("erro ao inserir evento na outbox: %w", err)
		}
	}

	// 6. Confirma a Transação Atômica
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("erro no commit: %w", err)
	}

	return nil
}

// CreateWalletTx insere uma nova carteira e, se houver saldo inicial, guarda o ledger e a outbox na mesma transação.
func (r *WalletRepository) CreateWalletTx(
	ctx context.Context,
	w domain.Wallet,
	ledger *domain.LedgerEntry,
	events []domain.OutboxEvent,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transacao: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Cria a Carteira
	_, err = tx.Exec(ctx, `
		INSERT INTO wallets (id, player_id, balance_cents, currency, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, w.ID(), w.PlayerID(), w.Balance().Cents(), w.Balance().Currency(), w.Version(), w.CreatedAt(), w.UpdatedAt())

	if err != nil {
		return fmt.Errorf("erro ao inserir carteira (pode já existir para este jogador e moeda): %w", err)
	}

	// 2. Se houver Ledger (saldo inicial > 0), cria o registo
	if ledger != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO wallet_ledger_entries 
			(id, wallet_id, transaction_id, direction, amount_cents, balance_before_cents, balance_after_cents, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, ledger.ID(), ledger.WalletID(), ledger.TransactionID(), ledger.Direction(),
			ledger.Amount().Cents(), ledger.BalanceBefore().Cents(), ledger.BalanceAfter().Cents(), ledger.CreatedAt())

		if err != nil {
			return fmt.Errorf("erro ao inserir ledger inicial: %w", err)
		}
	}

	// 3. Se houver Eventos, guarda na Outbox
	for _, ev := range events {
		_, err = tx.Exec(ctx, `
			INSERT INTO outbox (event_id, aggregate_id, event_type, payload, created_at) 
			VALUES ($1, $2, $3, $4, $5)
		`, ev.EventID, ev.AggregateID, ev.EventType, ev.Payload, ev.CreatedAt)
		if err != nil {
			return fmt.Errorf("erro ao inserir evento na outbox: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("erro no commit de criacao: %w", err)
	}
	return nil
}
