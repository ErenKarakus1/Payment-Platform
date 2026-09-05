package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID                  uuid.UUID `json:"id"`
	MerchantID          uuid.UUID `json:"merchant_id"`
	CustomerID          uuid.UUID `json:"customer_id"`
	AmountCents         int64     `json:"amount_cents"`
	RefundedAmountCents int64     `json:"refunded_amount_cents"`
	Currency            string    `json:"currency"`
	Status              string    `json:"status"`
	IdempotencyKey      uuid.UUID `json:"idempotency_key"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type CreatePaymentRequest struct {
	CustomerID  uuid.UUID `json:"customer_id"`
	AmountCents int64     `json:"amount_cents"`
	Currency    string    `json:"currency"`
}

func (r *CreatePaymentRequest) Normalize() {
	r.Currency = strings.ToUpper(strings.TrimSpace(r.Currency))
}
