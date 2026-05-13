package outbox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/childofemptiness/monitoring-system/contracts/events"
	"github.com/childofemptiness/url-monitor/internal/config"
	"github.com/childofemptiness/url-monitor/internal/ports"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeEventService struct {
	markPublishedGotCtx context.Context
	markForRetryGotCtx  context.Context
	markExhaustedGotCtx context.Context

	markPublishedGotEventID     uuid.UUID
	markPublishedGotPublishedAt time.Time
	markForRetryGotInput        ports.MarkFailedPublishInput
	markExhaustedGotInput       ports.MarkFailedPublishInput

	markForRetryErr  error
	markExhaustedErr error
	markPublishedErr error

	markForRetryCalled  bool
	markExhaustedCalled bool
	markPublishedCalled bool
}

func (f *fakeEventService) MarkForRetry(ctx context.Context, input ports.MarkFailedPublishInput) error {
	f.markForRetryGotCtx = ctx
	f.markForRetryCalled = true
	f.markForRetryGotInput = input
	return f.markForRetryErr
}

func (f *fakeEventService) MarkExhausted(ctx context.Context, input ports.MarkFailedPublishInput) error {
	f.markExhaustedGotCtx = ctx
	f.markExhaustedCalled = true
	f.markExhaustedGotInput = input
	return f.markExhaustedErr
}

func (f *fakeEventService) MarkPublished(ctx context.Context, eventID uuid.UUID, publishedAt time.Time) error {
	f.markPublishedGotCtx = ctx
	f.markPublishedGotEventID = eventID
	f.markPublishedGotPublishedAt = publishedAt
	f.markPublishedCalled = true
	return f.markPublishedErr
}

type fakeEventPublisher struct {
	gotCtx   context.Context
	gotEvent events.EventEnvelope

	publishCalled bool

	publishErr error
}

func (f *fakeEventPublisher) Publish(ctx context.Context, event events.EventEnvelope) error {
	f.gotCtx = ctx
	f.publishCalled = true
	f.gotEvent = event
	return f.publishErr
}

func TestEventProcessor_ProcessCanceled(t *testing.T) {
	service := &fakeEventService{}
	publisher := &fakeEventPublisher{}
	processor := NewEventProcessor(config.OutboxEventsConfig{}, service, publisher)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := processor.Process(ctx, Event{})
	require.Equal(t, err, context.Canceled)
	require.False(t, publisher.publishCalled)
	require.False(t, service.markPublishedCalled)
	require.False(t, service.markForRetryCalled)
	require.False(t, service.markExhaustedCalled)
}

func TestEventProcessor_ProcessTimeout(t *testing.T) {
	service := &fakeEventService{}
	publisher := &fakeEventPublisher{}
	processor := NewEventProcessor(config.OutboxEventsConfig{}, service, publisher)

	ctx, cancel := context.WithTimeout(context.Background(), -time.Second)
	defer cancel()

	err := processor.Process(ctx, Event{})
	require.Equal(t, err, context.DeadlineExceeded)
	require.False(t, publisher.publishCalled)
	require.False(t, service.markPublishedCalled)
	require.False(t, service.markForRetryCalled)
	require.False(t, service.markExhaustedCalled)
}

func TestEventProcessor_ProcessMarkPublished(t *testing.T) {
	service := &fakeEventService{}
	publisher := &fakeEventPublisher{}
	processor := NewEventProcessor(config.OutboxEventsConfig{}, service, publisher)

	ctx := context.Background()
	event := Event{
		EventID: uuid.New(),
	}
	mappedEvent := events.EventEnvelope{
		EventID:      event.EventID,
		EventType:    event.EventType,
		EventVersion: event.EventVersion,
		OccurredAt:   event.OccurredAt,
		Producer:     event.Producer,
		Payload:      event.Payload,
	}

	err := processor.Process(ctx, event)
	require.NoError(t, err)
	require.True(t, publisher.publishCalled)
	require.True(t, service.markPublishedCalled)
	require.False(t, service.markForRetryCalled)
	require.False(t, service.markExhaustedCalled)
	require.Equal(t, ctx, publisher.gotCtx)
	require.Equal(t, ctx, service.markPublishedGotCtx)
	require.Equal(t, mappedEvent, publisher.gotEvent)
	require.Equal(t, event.EventID, service.markPublishedGotEventID)
	require.NotEmpty(t, service.markPublishedGotPublishedAt)
}

func TestEventProcessor_ProcessMarkForRetry(t *testing.T) {
	service := &fakeEventService{}
	publisher := &fakeEventPublisher{}
	cfg := config.OutboxEventsConfig{
		MaxAttempts:  5,
		RetryBackoff: 1 * time.Second,
	}
	processor := NewEventProcessor(cfg, service, publisher)

	expectedErr := errors.New("publish error")
	publisher.publishErr = expectedErr

	ctx := context.Background()

	lastAttemptAt := time.Now().Add(-5 * time.Second)
	event := Event{
		EventID:       uuid.New(),
		AttemptsCount: 3,
		LastAttemptAt: &lastAttemptAt,
	}
	mappedEvent := events.EventEnvelope{
		EventID:      event.EventID,
		EventType:    event.EventType,
		EventVersion: event.EventVersion,
		OccurredAt:   event.OccurredAt,
		Producer:     event.Producer,
		Payload:      event.Payload,
	}

	err := processor.Process(ctx, event)
	require.NoError(t, err)
	require.True(t, publisher.publishCalled)
	require.True(t, service.markForRetryCalled)
	require.False(t, service.markExhaustedCalled)
	require.Equal(t, ctx, publisher.gotCtx)
	require.Equal(t, mappedEvent, publisher.gotEvent)
	require.Equal(t, ctx, service.markForRetryGotCtx)
	require.Equal(t, event.EventID, service.markForRetryGotInput.EventID)
	require.Equal(t, expectedErr.Error(), service.markForRetryGotInput.LastError)
	require.NotEmpty(t, service.markForRetryGotInput.LastAttemptAt)
}

func TestEventProcessor_ProcessMarkExhausted(t *testing.T) {
	service := &fakeEventService{}
	publisher := &fakeEventPublisher{}
	cfg := config.OutboxEventsConfig{
		MaxAttempts: 5,
	}

	processor := NewEventProcessor(cfg, service, publisher)

	expectedErr := errors.New("publish error")
	publisher.publishErr = expectedErr

	ctx := context.Background()

	lastAttemptAt := time.Now().Add(-5 * time.Second)
	event := Event{
		EventID:       uuid.New(),
		AttemptsCount: 4,
		LastAttemptAt: &lastAttemptAt,
		NextAttemptAt: lastAttemptAt.Add(3 * time.Second),
	}

	err := processor.Process(ctx, event)
	require.NoError(t, err)
	require.True(t, publisher.publishCalled)
	require.True(t, service.markExhaustedCalled)
	require.False(t, service.markForRetryCalled)
	require.Equal(t, ctx, publisher.gotCtx)
	require.Equal(t, ctx, service.markExhaustedGotCtx)
	require.Equal(t, event.EventID, service.markExhaustedGotInput.EventID)
	require.Equal(t, expectedErr.Error(), service.markExhaustedGotInput.LastError)
	require.Equal(t, event.NextAttemptAt, service.markExhaustedGotInput.NextAttemptAt)
	require.NotEmpty(t, service.markExhaustedGotInput.LastAttemptAt)
}

func TestEventProcessor_ProcessMarkPublishedError(t *testing.T) {
	service := &fakeEventService{}
	publisher := &fakeEventPublisher{}
	processor := NewEventProcessor(config.OutboxEventsConfig{}, service, publisher)

	expectedErr := errors.New("mark published error")
	service.markPublishedErr = expectedErr

	ctx := context.Background()
	event := Event{
		EventID: uuid.New(),
	}

	err := processor.Process(ctx, event)
	require.True(t, publisher.publishCalled)
	require.True(t, service.markPublishedCalled)
	require.False(t, service.markForRetryCalled)
	require.False(t, service.markExhaustedCalled)
	require.Equal(t, err, expectedErr)
	require.Equal(t, ctx, publisher.gotCtx)
	require.Equal(t, ctx, service.markPublishedGotCtx)
	require.Equal(t, event.EventID, service.markPublishedGotEventID)
}

func TestEventProcessor_ProcessMarkForRetryError(t *testing.T) {
	service := &fakeEventService{}
	publisher := &fakeEventPublisher{}
	cfg := config.OutboxEventsConfig{
		MaxAttempts:  5,
		RetryBackoff: 1 * time.Second,
	}
	processor := NewEventProcessor(cfg, service, publisher)

	expectedPublishErr := errors.New("publish error")
	publisher.publishErr = expectedPublishErr

	expectedMarkForRetryErr := errors.New("mark for retry error")
	service.markForRetryErr = expectedMarkForRetryErr

	ctx := context.Background()

	lastAttemptAt := time.Now().Add(-5 * time.Second)
	event := Event{
		EventID:       uuid.New(),
		AttemptsCount: 3,
		LastAttemptAt: &lastAttemptAt,
	}

	err := processor.Process(ctx, event)
	require.True(t, publisher.publishCalled)
	require.False(t, service.markPublishedCalled)
	require.True(t, service.markForRetryCalled)
	require.False(t, service.markExhaustedCalled)
	require.Equal(t, err, expectedMarkForRetryErr)
	require.Equal(t, ctx, publisher.gotCtx)
	require.Equal(t, ctx, service.markForRetryGotCtx)
	require.Equal(t, event.EventID, service.markForRetryGotInput.EventID)
	require.Equal(t, expectedPublishErr.Error(), service.markForRetryGotInput.LastError)
	require.NotEmpty(t, service.markForRetryGotInput.LastAttemptAt)
}

func TestEventProcessor_ProcessMarkExhaustedError(t *testing.T) {
	service := &fakeEventService{}
	publisher := &fakeEventPublisher{}
	cfg := config.OutboxEventsConfig{
		MaxAttempts:  5,
		RetryBackoff: 1 * time.Second,
	}
	processor := NewEventProcessor(cfg, service, publisher)

	expectedPublishErr := errors.New("publish error")
	publisher.publishErr = expectedPublishErr

	expectedMarkExhaustedErr := errors.New("mark exhausted error")
	service.markExhaustedErr = expectedMarkExhaustedErr

	ctx := context.Background()

	lastAttemptAt := time.Now().Add(-5 * time.Second)
	event := Event{
		EventID:       uuid.New(),
		AttemptsCount: 4,
		LastAttemptAt: &lastAttemptAt,
		NextAttemptAt: lastAttemptAt.Add(3 * time.Second),
	}

	err := processor.Process(ctx, event)
	require.True(t, publisher.publishCalled)
	require.False(t, service.markPublishedCalled)
	require.False(t, service.markForRetryCalled)
	require.True(t, service.markExhaustedCalled)
	require.Equal(t, err, expectedMarkExhaustedErr)
	require.Equal(t, ctx, publisher.gotCtx)
	require.Equal(t, ctx, service.markExhaustedGotCtx)
	require.Equal(t, event.EventID, service.markExhaustedGotInput.EventID)
	require.Equal(t, expectedPublishErr.Error(), service.markExhaustedGotInput.LastError)
	require.Equal(t, event.NextAttemptAt, service.markExhaustedGotInput.NextAttemptAt)
	require.NotEmpty(t, service.markExhaustedGotInput.LastAttemptAt)
}
