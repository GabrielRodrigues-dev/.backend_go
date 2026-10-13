package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/fx"

	"github.com/GabrielRodrigues-dev/backend_go/internal/domain"
	"github.com/GabrielRodrigues-dev/backend_go/internal/infrastructure/db"
	"github.com/GabrielRodrigues-dev/backend_go/internal/repository"
)

func main() {
	app := fx.New(
		fx.Provide(
			db.NewPostgresPool,
			repository.NewWalletRepository,
		),
		fx.Invoke(func(repo *repository.WalletRepository) {
			fmt.Println("\n=== PREPARANDO TESTE DE CONCORRÊNCIA ===")
			ctx := context.Background()

			// 1. criação de uma carteira com 100.00 BRL
			// GERANDO IDs DINÂMICOS PARA NÃO BATER NA REGRA DE UNICIDADE DO BANCO
			uniqueSuffix := time.Now().UnixNano()
			walletID := fmt.Sprintf("wallet-teste-%d", uniqueSuffix)
			playerID := fmt.Sprintf("player-%d", uniqueSuffix)

			initialMoney, _ := domain.NewMoneyFromString("100.00", "BRL")
			wallet := domain.NewWallet(walletID, playerID, initialMoney)

			// AGORA ESTAMOS TRATANDO O ERRO SE A CRIAÇÃO FALHAR
			err := repo.Create(ctx, wallet)
			if err != nil {
				fmt.Printf("Erro fatal ao criar carteira no banco: %v\n", err)
				return // Interrompe o programa aqui se der erro
			}

			fmt.Printf("Carteira criada com sucesso! ID: %s | Saldo: %s\n\n", wallet.ID(), wallet.Balance().Amount())

			// 2. Dispara DUAS apostas simultâneas de 80.00 BRL
			var wg sync.WaitGroup
			wg.Add(2)

			fmt.Println(">> Disparando Aposta A (80.00) e Aposta B (80.00) no mesmo milissegundo...")

			betAmount, _ := domain.NewMoneyFromString("80.00", "BRL")

			// Função que tenta fazer a aposta
			makeBet := func(betName string) {
				defer wg.Done()

				// A. Carrega a carteira do banco
				// TRATANDO ERRO DA BUSCA
				w, err := repo.FindByID(ctx, walletID)
				if err != nil {
					fmt.Printf("[%s] Erro ao buscar carteira no banco: %v\n", betName, err)
					return
				}

				// B. Tenta descontar 80.00
				err = w.Debit(betAmount)
				if err != nil {
					fmt.Printf("[%s] Rejeitada pelo domínio: %v\n", betName, err)
					return
				}

				// C. Cria o registro no extrato
				// O extrato garantindo um ID único para cada aposta simultânea
				ledgerEntry := domain.NewLedgerEntry(
					fmt.Sprintf("txn-%d-%s", time.Now().UnixNano(), betName),
					w.ID(), "ext-txn-123", "DEBIT", betAmount, initialMoney, w.Balance(),
				)

				// D. Tenta salvar a transação
				err = repo.ProcessTransaction(ctx, w.ID(), ledgerEntry, w.Version())
				if err != nil {
					fmt.Printf("[%s] Falhou ao salvar no banco: %v\n", betName, err)
				} else {
					fmt.Printf("[%s] SUCESSO! Aposta processada. Novo saldo: %s\n", betName, w.Balance().Amount())
				}
			}

			// Lançamos as duas em Goroutines (processamento paralelo)
			go makeBet("Aposta A")
			go makeBet("Aposta B")

			// Espera as duas terminarem
			wg.Wait()

			// 3. Resultado Final
			// Buscamos o resultado final sem ignorar erros
			finalWallet, err := repo.FindByID(ctx, walletID)
			if err != nil {
				fmt.Printf("\nErro ao buscar resultado final: %v\n", err)
				return
			}

			fmt.Printf("\n=== RESULTADO FINAL ===\n")
			fmt.Printf("Saldo Final da Carteira: %s (Versão %d)\n\n", finalWallet.Balance().Amount(), finalWallet.Version())
		}),
	)

	app.Run()
}
