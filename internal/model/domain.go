package model

import (
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID          int64     `json:"id"`
	EventId     uuid.UUID `json:"event_id"`
	Payload     string    `json:"payload"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	NextRetryAt time.Time `json:"next_retry_at"`
	Attempts    int       `json:"attempts"`
}
