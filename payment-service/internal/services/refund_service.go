package services

import (
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/google/uuid"
)

func createRefund(paymentID uuid.UUID, merchantID uuid.UUID, refundAmount int64) models.Refund {
	return models.Refund{
		ID:          uuid.New(),
		PaymentID:   paymentID,
		MerchantID:  merchantID,
		AmountCents: refundAmount,
	}
}
