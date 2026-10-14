package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/fx"

	"github.com/GabrielRodrigues-dev/backend_go/internal/handler"
)

// Configura as rotas e associa o servidor ao ciclo de vida do Uber Fx.
func StartHTTPServer(lc fx.Lifecycle, walletHandler *handler.WalletHandler) {
	mux := http.NewServeMux()

	// Registamos a nossa rota enrolada no Middleware de Autenticação
	mux.HandleFunc("/bet", AuthMiddleware(walletHandler.PlaceBet))
	mux.HandleFunc("/wallets", AuthMiddleware(walletHandler.CreateWallet))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			fmt.Println("Iniciando Servidor HTTP na porta 8080...")
			// O servidor HTTP trava a execução.
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					fmt.Printf("Erro no servidor HTTP: %v\n", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Encerrando Servidor HTTP (Graceful Shutdown)...")
			// Espera as requisições em andamento terminarem antes de desligar.
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return srv.Shutdown(shutdownCtx)
		},
	})
}
