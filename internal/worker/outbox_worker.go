package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

type OutboxWorker struct {
	db *pgxpool.Pool
}

func NewOutboxWorker(db *pgxpool.Pool) *OutboxWorker {
	return &OutboxWorker{db: db}
}

// StartOutboxWorker registra a Goroutine de envio de eventos no ciclo de vida da aplicação.
func StartOutboxWorker(lc fx.Lifecycle, worker *OutboxWorker) {
	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(c context.Context) error {
			fmt.Println("Iniciando Outbox Publisher Worker...")
			go worker.publishLoop(ctx)
			return nil
		},
		OnStop: func(c context.Context) error {
			fmt.Println("Encerrando Outbox Publisher Worker...")
			cancel()
			return nil
		},
	})
}

func (w *OutboxWorker) publishLoop(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processPendingEvents(ctx)
		}
	}
}

func (w *OutboxWorker) processPendingEvents(ctx context.Context) {
	// 1. Busca eventos não publicados (Limitado a 50 por vez para não sobrecarregar a memória)
	// O FOR UPDATE SKIP LOCKED é essencial para concorrência
	rows, err := w.db.Query(ctx, `
		SELECT event_id, event_type, payload 
		FROM outbox 
		WHERE published = false 
		ORDER BY created_at ASC 
		LIMIT 50 
		FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		log.Printf("Erro ao buscar eventos na outbox: %v", err)
		return
	}
	defer rows.Close()

	var eventsToMark []string

	for rows.Next() {
		var eventID, eventType string
		var payload []byte
		if err := rows.Scan(&eventID, &eventType, &payload); err != nil {
			continue
		}

		// 2. Simula a publicação em um serviço externo (como SNS, EventBridge ou Kafka)
		fmt.Printf("\n>> [OUTBOX PUBLISHER] Publicando Evento: %s | ID: %s\n", eventType, eventID)

		eventsToMark = append(eventsToMark, eventID)
	}

	// 3. Marca os eventos como publicados no banco de dados
	if len(eventsToMark) > 0 {
		_, err = w.db.Exec(ctx, `UPDATE outbox SET published = true WHERE event_id = ANY($1)`, eventsToMark)
		if err != nil {
			log.Printf("Erro ao marcar eventos da outbox como publicados: %v", err)
		} else {
			fmt.Printf(">> [OUTBOX PUBLISHER] %d eventos marcados como publicados.\n", len(eventsToMark))
		}
	}
}
