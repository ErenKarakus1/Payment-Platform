package services

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func createRefund(paymentID uuid.UUID, merchantID uuid.UUID, refundAmount int64) models.Refund {
	return models.Refund{
		ID:          uuid.New(),
		PaymentID:   paymentID,
		MerchantID:  merchantID,
		AmountCents: refundAmount,
	}
}

func GetAllRefundsByPaymentID(ctx context.Context, pool *pgxpool.Pool, paymentID uuid.UUID, merchantID uuid.UUID) ([]models.Refund, error) {
	payment, err := repository.GetPaymentByID(ctx, pool, paymentID, merchantID)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return []models.Refund{}, repository.ErrPaymentNotFound
		}
		return []models.Refund{}, ErrInternalServerError
	}
	if !(payment.Status == models.PaymentStatusSucceeded || payment.Status == models.PaymentStatusPartiallyRefunded || payment.Status == models.PaymentStatusRefunded) {
		return []models.Refund{}, errors.New("payment is not refundable")
	}
	refunds, err := repository.GetAllRefundsByPaymentID(ctx, pool, merchantID, paymentID)
	if err != nil {
		return []models.Refund{}, ErrInternalServerError
	}
	return refunds, nil
}
