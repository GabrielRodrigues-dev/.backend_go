package main

import (
	"go.uber.org/fx"

	myhttp "github.com/GabrielRodrigues-dev/backend_go/internal/Infrastructure/http"
	"github.com/GabrielRodrigues-dev/backend_go/internal/handler"
	"github.com/GabrielRodrigues-dev/backend_go/internal/infrastructure/db"
	"github.com/GabrielRodrigues-dev/backend_go/internal/repository"
	"github.com/GabrielRodrigues-dev/backend_go/internal/worker"
)

func main() {
	app := fx.New(
		// PROVIDE: Ensina o Fx a construir as ferramentas
		fx.Provide(
			db.NewPostgresPool,
			db.NewSQSClient, // <--- Cliente AWS
			repository.NewWalletRepository,
			handler.NewWalletHandler,
			worker.NewSQSWorker, // <--- Construtor do Worker
		),
		// INVOKE: Diz ao Fx o que deve arrancar
		fx.Invoke(
			myhttp.StartHTTPServer, // Isto vai acionar o OnStart que regista a porta 8080
			worker.StartWorker,     // <--- Inicia a Goroutine do SQS no arranque
		),
	)

	app.Run()
}
