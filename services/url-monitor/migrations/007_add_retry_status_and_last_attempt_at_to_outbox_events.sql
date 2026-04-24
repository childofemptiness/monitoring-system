ALTER TABLE outbox_events
ADD COLUMN last_attempt_at TIMESTAMPTZ;

ALTER TABLE outbox_events
DROP CONSTRAINT chk_outbox_events_status;

ALTER TABLE outbox_events
ADD CONSTRAINT chk_outbox_events_status
    CHECK (status IN ('pending', 'processing', 'retry', 'published', 'exhausted'));

DROP INDEX IF EXISTS idx_outbox_events_status_next_attempt_at;

CREATE INDEX idx_outbox_events_status_pending
ON outbox_events (status)
WHERE status = 'pending';

CREATE INDEX idx_outbox_events_status_retry_next_attempt_at
ON outbox_events (status, next_attempt_at)
WHERE status = 'retry';
