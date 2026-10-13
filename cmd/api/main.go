package main

import (
	"go.uber.org/fx"

	myhttp "github.com/GabrielRodrigues-dev/backend_go/internal/Infrastructure/http"
	"github.com/GabrielRodrigues-dev/backend_go/internal/handler"
	"github.com/GabrielRodrigues-dev/backend_go/internal/infrastructure/db"
	"github.com/GabrielRodrigues-dev/backend_go/internal/repository"
)

func main() {
	app := fx.New(
		// PROVIDE: Ensina o Fx a construir as ferramentas
		fx.Provide(
			db.NewPostgresPool,
			repository.NewWalletRepository,
			handler.NewWalletHandler,
		),
		// INVOKE: Diz ao Fx o que deve arrancar
		fx.Invoke(
			myhttp.StartHTTPServer, // Isto vai acionar o OnStart que regista a porta 8080
		),
	)

	app.Run()
}
