package rabbitmq_module

import amqp "github.com/rabbitmq/amqp091-go"

type acknowledger interface {
	Body() []byte
	Ack(multiple bool) error
	Nack(multiple, requeue bool) error
}
type rabbitDelivery struct {
	msg amqp.Delivery
}

func (d *rabbitDelivery) Body() []byte {
	return d.msg.Body
}

func (d *rabbitDelivery) Ack(multiple bool) error {
	return d.msg.Ack(multiple)
}

func (d *rabbitDelivery) Nack(multiple, requeue bool) error {
	return d.msg.Nack(multiple, requeue)
}
