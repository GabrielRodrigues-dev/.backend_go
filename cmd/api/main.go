package main

import (
	"context"
	"fmt"

	"github.com/GabrielRodrigues-dev/backend_go/internal/infrastructure/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		//PROVIDE: Ensinamos o Fx a criar o pool do PostgreSQL.
		fx.Provide(
			db.NewPostgresPool,
		),

		//INVOKE: O que a aplicação deve fazer com o que construiu.
		//Pedimos a instância do *pgxpool.Pool apenas para obrigar o Fx a criá-la e testá-la.
		fx.Invoke(func(pool *pgxpool.Pool) {
			// Se o código chegou aqui, o OnStart do Lifecycle correu bem e a base de dados está pronta!
			var dbName string
			err := pool.QueryRow(context.Background(), "SELECT current_database();").Scan(&dbName)
			if err == nil {
				fmt.Printf("Sucesso! Conectado à base de dados: %s\n", dbName)
			}
		}),
	)

	app.Run()
}
