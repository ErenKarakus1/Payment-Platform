package models

import (
	"time"

	"github.com/google/uuid"
)

type Refund struct {
	ID          uuid.UUID `json:"id"`
	PaymentID   uuid.UUID `json:"payment_id"`
	MerchantID  uuid.UUID `json:"merchant_id"`
	AmountCents int64     `json:"amount_cents"`
	CreatedAt   time.Time `json:"created_at"`
}
