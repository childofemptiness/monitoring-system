package check

import (
	"context"
	"time"
	"url-monitor/internal/events"
	"url-monitor/internal/monitor"

	"github.com/google/uuid"
)

type CheckRepository interface {
	CompleteCheck(ctx context.Context, input CreateCheckWithEventInput) error
}

type CheckStoreService struct {
	repo CheckRepository
}

type CreateCheckWithEventInput struct {
	MonitorID      int64
	URL            string
	Status         monitor.MonitorCheckStatus
	HTTPStatusCode int16
	ErrorMessage   string
	ErrorKind      monitor.CheckErrorKind
	ResponseTimeMS int64
	StartedAt      time.Time
	FinishedAt     time.Time

	EventID      uuid.UUID
	EventType    events.EventType
	EventVersion int
	OccurredAt   time.Time
	Producer     events.EventProducer

	NextCheckAt time.Time
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

	input := CreateCheckWithEventInput{
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
		EventVersion: events.EventVersionURLChecked,
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
