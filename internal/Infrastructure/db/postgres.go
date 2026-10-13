package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

// NewPostgresPool é o construtor que o Uber Fx vai utilizar.
func NewPostgresPool(lc fx.Lifecycle) (*pgxpool.Pool, error) {
	// O endereço da base de dados que subimos no Docker na Etapa 3
	connString := "postgres://jungle:password@localhost:5432/betting_db?sslmode=disable"

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter string de conexão: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar pool do postgres: %w", err)
	}

	// O fx.Lifecycle gere o arranque e o encerramento dos componentes.
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("Conectando ao PostgreSQL...")
			return pool.Ping(ctx) // Testa a ligação ao arrancar
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Encerrando conexão com o PostgreSQL...")
			pool.Close() // Fecha o cofre de forma segura quando a aplicação parar
			return nil
		},
	})

	return pool, nil
}
