package domain

import (
	"encoding/json"
	"time"
)

// Representa um evento que será publicado de forma assíncrona.
type OutboxEvent struct {
	EventID     string
	AggregateID string
	EventType   string
	Payload     []byte // O JSON imutável exigido pelo desafio
	CreatedAt   time.Time
}

// Empacota qualquer struct de dados (payload) num evento em JSON.
func NewOutboxEvent(eventID, aggregateID, eventType string, payload any) (OutboxEvent, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return OutboxEvent{}, err
	}

	return OutboxEvent{
		EventID:     eventID,
		AggregateID: aggregateID,
		EventType:   eventType,
		Payload:     data,
		CreatedAt:   time.Now().UTC(),
	}, nil
}
