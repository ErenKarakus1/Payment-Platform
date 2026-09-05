package services

import (
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/google/uuid"
)

const PaymentStatusPending = "pending"

func CreatePaymentFromCreatePaymentRequest(r models.CreatePaymentRequest, merchantID uuid.UUID, idempotencyKey uuid.UUID) models.Payment {
	return models.Payment{
		ID:             uuid.New(),
		MerchantID:     merchantID,
		CustomerID:     r.CustomerID,
		AmountCents:    r.AmountCents,
		Currency:       r.Currency,
		Status:         PaymentStatusPending,
		IdempotencyKey: idempotencyKey,
	}
}
