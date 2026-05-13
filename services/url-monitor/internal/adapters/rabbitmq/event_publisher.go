package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/childofemptiness/monitoring-system/contracts/events"
)

type MessagePublisher interface {
	Publish(ctx context.Context, body []byte) error
}
type RabbitPublisherAdapter struct {
	publisher MessagePublisher
}

func NewRabbitPublisherAdapter(publisher MessagePublisher) *RabbitPublisherAdapter {
	return &RabbitPublisherAdapter{publisher: publisher}
}

func (a *RabbitPublisherAdapter) Publish(ctx context.Context, event events.EventEnvelope) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return a.publisher.Publish(ctx, body)
}
