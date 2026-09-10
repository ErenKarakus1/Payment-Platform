package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ErenKarakus1/Payment-Platform/notification-service/internal/mail"
	"github.com/ErenKarakus1/Payment-Platform/notification-service/internal/models"
	"github.com/ErenKarakus1/Payment-Platform/notification-service/internal/services"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewReader(broker string, topic string, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{broker},
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

func (c *Consumer) Consume(ctx context.Context, sender *mail.Sender) error {
	for {
		message, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		var event models.PaymentEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			log.Printf("failed to unmarshal payment event: %v", event)
			continue
		}
		log.Printf("recieved event: type=%s payment_id=%s", event.EventType, event.PaymentID)
		services.HandlePaymentEvent(event, sender)
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
