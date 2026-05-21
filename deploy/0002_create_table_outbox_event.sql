-- +goose Up
CREATE TABLE IF NOT EXISTS outbox_event
(
    id            INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id      uuid      NOT NULL,
    payload       TEXT      NOT NULL,
    status        TEXT      NOT NULL DEFAULT 'PROCESSING' CHECK (status in ('PROCESSING', 'FAILED', 'SUCCESS')),
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    next_retry_at TIMESTAMP NOT NULL DEFAULT NOW(),
    attempts      INTEGER            DEFAULT 0 CHECK (attempts >= 0)
);

CREATE INDEX idx_outbox_status_retry ON outbox_event (status, next_retry_at);

-- +goose Down
DROP TABLE IF EXISTS outbox_event;
