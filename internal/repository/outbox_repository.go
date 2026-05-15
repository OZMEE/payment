package repository

import (
	"context"
	json2 "encoding/json"
	"fmt"
	"payment/internal/appers"
	"payment/internal/model"
	"payment/pkg/db"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	ProcessingStatus = "PROCESSING"
	SuccessStatus    = "SUCCESS"
	FailedStatus     = "FAILED"
)

type OutboxRepository interface {
	GetProcessingOutboxes(ctx context.Context, tx *sqlx.Tx, limitEvents int) ([]*model.OutboxEvent, error)
	CreateOutbox(ctx context.Context, tx *sqlx.Tx, payload any) error
	UpdateOutboxes(ctx context.Context, tx *sqlx.Tx, events []*model.OutboxEvent) error
	BeginTransaction() (*sqlx.Tx, error)
}

type OutboxRepositoryImpl struct {
	db *db.Database
}

func NewOutboxRepositoryImpl(db *db.Database) *OutboxRepositoryImpl {
	return &OutboxRepositoryImpl{db: db}
}

func (r *OutboxRepositoryImpl) BeginTransaction() (*sqlx.Tx, error) {
	tx, err := r.db.BeginTransaction()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r *OutboxRepositoryImpl) CreateOutbox(ctx context.Context, tx *sqlx.Tx, payload any) error {
	const op = "OutboxRepositoryImpl.CreateOutbox"

	query := `INSERT INTO outbox_event (event_id, payload) VALUES ($1, $2)`

	json, err := json2.Marshal(payload)
	if err != nil {
		return appers.ErrParseJson.Builder().Msg(err.Error()).Op(op).Build()
	}

	_, err = tx.QueryContext(ctx, query, uuid.New().String(), json)
	if err != nil {
		return appers.ErrSqlExecutions.Builder().Msg(err.Error()).Op(op).Build()
	}

	return nil
}

func (r *OutboxRepositoryImpl) GetProcessingOutboxes(ctx context.Context, tx *sqlx.Tx, limitEvents int) ([]*model.OutboxEvent, error) {
	const op = "OutboxRepositoryImpl.GetProcessingOutboxes"

	query := `SELECT id, event_id, payload, status, created_at, next_retry_at, attempts 
			  FROM outbox_event 
			  WHERE status = $1 AND next_retry_at <= NOW() 
			  ORDER BY created_at ASC 
			  LIMIT $2
			  FOR UPDATE SKIP LOCKED`

	rows, err := tx.QueryContext(ctx, query, ProcessingStatus, limitEvents)
	if err != nil {
		return nil, appers.ErrSqlExecutions.Builder().Msg(err.Error()).Op(op).Build()
	}

	if rows == nil {
		return nil, appers.ErrSqlExecutions.Builder().Op(op).Build()
	}

	events := make([]*model.OutboxEvent, 0)
	for rows.Next() {
		var outboxEvent model.OutboxEvent
		err := rows.Scan(&outboxEvent.ID, &outboxEvent.EventId, &outboxEvent.Payload, &outboxEvent.Status,
			&outboxEvent.CreatedAt, &outboxEvent.NextRetryAt, &outboxEvent.Attempts)
		if err != nil {
			return nil, appers.ErrSqlExecutions.Builder().Msg(err.Error()).Op(op).Build()
		}
		events = append(events, &outboxEvent)
	}

	return events, nil
}

func (r *OutboxRepositoryImpl) UpdateOutboxes(ctx context.Context, tx *sqlx.Tx, events []*model.OutboxEvent) error {
	const op = "OutboxRepositoryImpl.UpdateOutboxes"

	valueStrings := make([]string, len(events))
	args := make([]interface{}, 0, len(events)*4)
	counter := 1

	for i, event := range events {
		valueStrings[i] = fmt.Sprintf("($%d::bigint, $%d::text, $%d::timestamp, $%d::int)", counter, counter+1, counter+2, counter+3)
		args = append(args,
			event.ID,
			event.Status,
			event.NextRetryAt,
			event.Attempts)
		counter += 4
	}

	query := fmt.Sprintf(`
		WITH updates (id, status, next_retry_at, attempts) AS (
			VALUES %s
		)
		UPDATE outbox_event
		SET 
			status = updates.status, 
			next_retry_at = updates.next_retry_at, 
			attempts = updates.attempts
		FROM updates
		WHERE outbox_event.id = updates.id
		`, strings.Join(valueStrings, ","))

	_, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return appers.ErrSqlExecutions.Builder().Msg(err.Error()).Op(op).Build()
	}

	return nil
}
