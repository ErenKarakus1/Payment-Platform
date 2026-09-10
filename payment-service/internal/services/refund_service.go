package services

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/kafka"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/repository"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/validations"
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

func RefundPayment(ctx context.Context, pool *pgxpool.Pool, producer *kafka.Producer, merchantID uuid.UUID, paymentID uuid.UUID, refundRequest models.RefundRequest) (models.Payment, error) {
	if err := validations.ValidateRefundRequest(refundRequest); err != nil {
		return models.Payment{}, err
	}
	payment, err := repository.GetPaymentByID(ctx, pool, paymentID, merchantID)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return models.Payment{}, err
		}
		return models.Payment{}, ErrInternalServerError
	}
	if !(payment.Status == models.PaymentStatusSucceeded || payment.Status == models.PaymentStatusPartiallyRefunded) {
		if payment.Status == models.PaymentStatusRefunded {
			return models.Payment{}, errors.New("refund amount exceeds remaining refundable amount")
		}
		return models.Payment{}, errors.New("payment status is not compatible with refunds")
	}
	if (payment.AmountCents - payment.RefundedAmountCents) < refundRequest.AmountCents {
		return models.Payment{}, errors.New("refund amount exceeds remaining refundable amount")
	}
	var targetStatus string
	if payment.RefundedAmountCents+refundRequest.AmountCents == payment.AmountCents {
		targetStatus = models.PaymentStatusRefunded
	} else {
		targetStatus = models.PaymentStatusPartiallyRefunded
	}
	refund := createRefund(paymentID, merchantID, refundRequest.AmountCents)
	updatedPayment, err := repository.RefundPayment(ctx, pool, paymentID, merchantID, payment.RefundedAmountCents, payment.RefundedAmountCents+refundRequest.AmountCents, targetStatus, refund)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return models.Payment{}, err
		}
		return models.Payment{}, ErrInternalServerError
	}
	customer, err := repository.GetCustomerByID(ctx, pool, merchantID, updatedPayment.CustomerID)
	if err != nil {
		return models.Payment{}, ErrInternalServerError
	}
	err = producer.PublishPaymentEvent(ctx, kafka.PaymentEvent{
		EventType:     EventPaymentRefunded,
		PaymentID:     updatedPayment.ID,
		MerchantID:    merchantID,
		CustomerID:    updatedPayment.CustomerID,
		CustomerEmail: customer.Email,
		CustomerName:  customer.Name,
		AmountCents:   refund.AmountCents,
		Currency:      updatedPayment.Currency,
	})
	if err != nil {
		return models.Payment{}, ErrKafkaPublishEvent
	}
	return updatedPayment, nil
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
