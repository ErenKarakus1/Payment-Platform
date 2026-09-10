package kafka

import (
	"context"
	"encoding/json"
	"errors"

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
	EventType     string    `json:"event_type"`
	PaymentID     uuid.UUID `json:"payment_id"`
	MerchantID    uuid.UUID `json:"merchant_id"`
	CustomerID    uuid.UUID `json:"customer_id"`
	CustomerEmail string    `json:"customer_email"`
	AmountCents   int64     `json:"amount_cents"`
	Currency      string    `json:"currency"`
}

func (p *Producer) PublishPaymentEvent(ctx context.Context, event PaymentEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return errors.New("couldnt marshal event")
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
