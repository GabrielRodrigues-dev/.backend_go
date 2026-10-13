package repository

import (
	"context"
	"fmt"

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
