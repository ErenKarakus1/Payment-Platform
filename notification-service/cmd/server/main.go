package main

import (
	"context"
	"log"

	"github.com/ErenKarakus1/Payment-Platform/notification-service/internal/config"
	"github.com/ErenKarakus1/Payment-Platform/notification-service/internal/kafka"
	"github.com/ErenKarakus1/Payment-Platform/notification-service/internal/mail"
)

func main() {
	cfg := config.LoadConfig()
	ctx := context.Background()

	consumer := kafka.NewReader(
		"localhost:9092",
		"payment.events",
		"notification-service",
	)
	defer consumer.Close()
	log.Println("Kafka consumer initialized")

	sender := mail.NewSender(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
	)

	if err := consumer.Consume(ctx, sender); err != nil {
		log.Fatal(err)
	}
}
