package check

import (
	"context"
	"time"
	"url-monitor/internal/monitor"
	"url-monitor/internal/outbox"
	"url-monitor/internal/ports"

	"github.com/childofemptiness/monitoring-system/contracts/events"
	"github.com/google/uuid"
)

type CheckRepository interface {
	CompleteCheck(ctx context.Context, input ports.CreateCheckWithEventInput) error
}

type CheckStoreService struct {
	repo CheckRepository
}

func NewCheckStoreService(repo CheckRepository) *CheckStoreService {
	return &CheckStoreService{repo: repo}
}

func (c *CheckStoreService) SaveCheckResult(
	ctx context.Context,
	check monitor.MonitorCheck,
	nextCheckAt time.Time,
	monitorURL string,
) error {
	eventID, err := getNewEventID()
	if err != nil {
		return err
	}

	input := ports.CreateCheckWithEventInput{
		MonitorID:      check.MonitorID,
		URL:            monitorURL,
		Status:         check.Status,
		HTTPStatusCode: check.HTTPStatusCode,
		ErrorMessage:   check.ErrorMessage,
		ErrorKind:      check.ErrorKind,
		ResponseTimeMS: check.ResponseTimeMS,
		StartedAt:      check.StartedAt,
		FinishedAt:     check.FinishedAt,

		EventID:      eventID,
		EventType:    events.EventTypeURLChecked,
		EventVersion: outbox.EventVersionURLChecked,
		Producer:     events.EventProducerURLMonitor,
		OccurredAt:   check.FinishedAt,

		NextCheckAt: nextCheckAt,
	}

	return c.repo.CompleteCheck(ctx, input)
}

func getNewEventID() (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
