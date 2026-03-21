CREATE TABLE IF NOT EXISTS monitor_checks (
    id               BIGSERIAL PRIMARY KEY,
    monitor_id       BIGINT NOT NULL,
    status           VARCHAR(32) NOT NULL,
    http_status_code SMALLINT,
    error_message    VARCHAR(255),
    response_time_ms INTEGER,
    started_at       TIMESTAMPTZ NOT NULL,
    finished_at      TIMESTAMPTZ,

    CONSTRAINT fk_monitor_checks_monitor_id
    FOREIGN KEY (monitor_id) REFERENCES monitors(id)
)
