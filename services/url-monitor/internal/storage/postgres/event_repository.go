package postgres

import (
	"context"
	"time"

	"github.com/childofemptiness/url-monitor/internal/outbox"
	"github.com/childofemptiness/url-monitor/internal/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	db *pgxpool.Pool
}

func (r *EventRepository) MarkForRetry(ctx context.Context, input ports.MarkFailedPublishInput) error {
	//TODO implement me
	panic("implement me")
}

func (r *EventRepository) MarkExhausted(ctx context.Context, input ports.MarkFailedPublishInput) error {
	//TODO implement me
	panic("implement me")
}

func (r *EventRepository) MarkPublished(ctx context.Context, publishedAt time.Time, eventID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) ClaimNextBatch(ctx context.Context, input ports.ClaimNextBatchInput) ([]outbox.Event, error) {
	query := `
		SELECT
		    event_id,
		    event_type,
		    event_version,
		    status,
		    producer,
		    attempts_count,
		    last_error,
		    occurred_at,
		    created_at,
		    last_attempt_at,
		    next_attempt_at,
		    processing_started_at,
		    published_at,
		    payload
		FROM outbox_events
		WHERE (status = 'pending' AND next_attempt_at <= $1)
		OR (status = 'processing' AND $1 - processing_started_at >= $2)
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, input.Now, input.ProcessingTimeout, input.FetchEventsLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claimed []outbox.Event
	for rows.Next() {
		var event outbox.Event
		if err := rows.Scan(
			&event.EventID,
			&event.EventType,
			&event.EventVersion,
			&event.Status,
			&event.Producer,
			&event.AttemptsCount,
			&event.LastError,
			&event.OccurredAt,
			&event.CreatedAt,
			&event.LastAttemptAt,
			&event.NextAttemptAt,
			&event.ProcessingStartedAt,
			&event.PublishedAt,
			&event.Payload,
		); err != nil {
			return nil, err
		}

		claimed = append(claimed, event)
	}

	return claimed, nil
}
