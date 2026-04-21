CREATE TABLE IF NOT EXISTS outbox_events (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID NOT NULL UNIQUE,
    event_type VARCHAR(32) NOT NULL,
    event_version SMALLINT NOT NULL,
    status VARCHAR(32) NOT NULL,
    producer VARCHAR(64) NOT NULL,
    attempts_count SMALLINT NOT NULL DEFAULT 0,
    last_error TEXT,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processing_started_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,

    CONSTRAINT chk_outbox_events_status
        CHECK (status IN ('pending', 'processing', 'exhausted', 'published'))
);

CREATE INDEX idx_outbox_events_status_next_attempt_at
ON outbox_events (status, next_attempt_at)
WHERE status = 'pending';

CREATE INDEX idx_outbox_events_status_processing_started_at
ON outbox_events (status, processing_started_at)
WHERE status = 'processing';
