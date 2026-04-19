CREATE TABLE monitors (
    id BIGSERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    interval_seconds INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
