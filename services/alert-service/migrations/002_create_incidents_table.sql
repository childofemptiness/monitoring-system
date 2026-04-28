CREATE TABLE IF NOT EXISTS incidents (
    id BIGSERIAL PRIMARY KEY,
    opened_by_event_id UUID NOT NULL REFERENCES inbox_events(event_id),
    closed_by_event_id UUID REFERENCES inbox_events(event_id),
    target_key TEXT NOT NULL,
    status VARCHAR(32) NOT NULL,
    opened_at TIMESTAMPTZ NOT NULL,
    closed_at TIMESTAMPTZ,

    CONSTRAINT chk_incidents_lifecycle
    CHECK (
        (
            status = 'opened'
            AND opened_by_event_id IS NOT NULL
            AND closed_by_event_id IS NULL
            AND opened_at IS NOT NULL
            AND closed_at IS NULL
        )
            OR
        (
            status = 'closed'
            AND opened_by_event_id IS NOT NULL
            AND closed_by_event_id IS NOT NULL
            AND opened_at IS NOT NULL
            AND closed_at IS NOT NULL
            AND closed_at >= opened_at
        )
    )
);

CREATE UNIQUE INDEX uniq_incidents_opened_target
ON incidents (target_key)
WHERE status = 'opened';
