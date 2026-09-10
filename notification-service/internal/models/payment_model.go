package models

import "github.com/google/uuid"

type PaymentEvent struct {
	EventType     string    `json:"event_type"`
	PaymentID     uuid.UUID `json:"payment_id"`
	MerchantID    uuid.UUID `json:"merchant_id"`
	CustomerID    uuid.UUID `json:"customer_id"`
	CustomerEmail string    `json:"customer_email"`
	CustomerName  string    `json:"customer_name"`
	AmountCents   int64     `json:"amount_cents"`
	Currency      string    `json:"currency"`
}
