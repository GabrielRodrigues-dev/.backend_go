package main

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// A string de conexão aponta para o container do Docker
	dbURL := "postgres://jungle:password@localhost:5432/betting_db?sslmode=disable"

	// O source aponta para a pasta onde guarda o ficheiros SQL
	sourceURL := "file://migrations"

	log.Printf("Buscando migrations em: %s\n", sourceURL)

	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		log.Fatalf("Erro ao iniciar migration: %v", err)
	}

	// Executa todas as migrations pendentes (Up)
	log.Println("Aplicando migrations...")
	err = m.Up()

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Erro ao aplicar migrations: %v", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("Nenhuma nova migration para aplicar.")
	} else {
		log.Println("Migrations aplicadas com sucesso! As tabelas foram criadas.")
	}
}
