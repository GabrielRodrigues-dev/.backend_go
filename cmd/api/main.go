package main

import (
	"context"
	"fmt"

	"github.com/GabrielRodrigues-dev/backend_go/internal/domain"
	"github.com/GabrielRodrigues-dev/backend_go/internal/infrastructure/db"
	"github.com/GabrielRodrigues-dev/backend_go/internal/repository"

	//"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		//PROVIDE: Ensinamos o Fx a criar o pool do PostgreSQL.
		fx.Provide(
			db.NewPostgresPool,             // cria a conexão com o banco de dados
			repository.NewWalletRepository, // cria o repositotio
		),

		//INVOKE: O que a aplicação deve fazer com o que construiu.
		//Pedimos a instância do *pgxpool.Pool apenas para obrigar o Fx a criá-la e testá-la.
		fx.Invoke(func(repo *repository.WalletRepository) {
			fmt.Println("Testando a inserção de uma carteira...")

			// cria um dinheiro e uma carteira na memória.
			initialMoney, _ := domain.NewMoneyFromString("100.00", "BRL")
			wallet := domain.NewWallet("wallet-teste-01", "player-123", initialMoney)

			// Usando context.Background() porque não temos uma requisição HTTP ainda
			err := repo.Create(context.Background(), wallet)
			if err != nil {
				fmt.Printf("Falha: a carteira já existe ou ocorreu um erro: %v\n", err)
			} else {
				fmt.Println("Sucesso! Carteira salva no PostgreSQL.")
			}
		}),
	)

	app.Run()
}
