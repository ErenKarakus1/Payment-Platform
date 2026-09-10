package kafka

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(broker string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

type PaymentEvent struct {
	EventType  string    `json:"event_type"`
	PaymentID  uuid.UUID `json:"payment_id"`
	MerchantID uuid.UUID `json:"merchant_id"`
}

func (p *Producer) PublishPaymentEvent(ctx context.Context, event PaymentEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(
		ctx,
		kafka.Message{
			Key:   []byte(event.PaymentID.String()),
			Value: data,
		},
	)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
