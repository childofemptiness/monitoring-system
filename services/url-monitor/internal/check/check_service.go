package check

import (
	"context"
	"time"
	"url-monitor/internal/events"
	"url-monitor/internal/monitor"

	"github.com/google/uuid"
)

type CheckRepository interface {
	CompleteCheck(ctx context.Context, check monitor.MonitorCheck, event events.URLChecked, nextCheckAt time.Time) error
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

	urlCheckedEvent := events.URLChecked{
		EventID:      eventID,
		EventType:    events.EventTypeURLChecked,
		EventVersion: events.EventVersionURLChecked,
		OccurredAt:   check.FinishedAt,
		Producer:     events.EventProducerURLMonitor,
		Payload: events.Payload{
			MonitorID: check.MonitorID,
			URL:       monitorURL,
			Status:    check.Status,
			CheckedAt: check.FinishedAt,
		},
	}

	if check.Status != monitor.MonitorCheckStatusError {
		urlCheckedEvent.Payload.HTTPStatusCode = &check.HTTPStatusCode
	} else if check.ErrorKind != "" {
		urlCheckedEvent.Payload.ErrorKind = &check.ErrorKind
	}

	return c.repo.CompleteCheck(ctx, check, urlCheckedEvent, nextCheckAt)
}

func getNewEventID() (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
