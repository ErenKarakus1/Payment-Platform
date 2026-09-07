package services

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/repository"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/validations"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func createPaymentFromCreatePaymentRequest(r models.CreatePaymentRequest, merchantID uuid.UUID, idempotencyKey uuid.UUID) models.Payment {
	return models.Payment{
		ID:             uuid.New(),
		MerchantID:     merchantID,
		CustomerID:     r.CustomerID,
		AmountCents:    r.AmountCents,
		Currency:       r.Currency,
		Status:         models.PaymentStatusPending,
		IdempotencyKey: idempotencyKey,
	}
}

var ErrInvalidStatusTransition = errors.New("invalid status transition")

func CreatePayment(ctx context.Context, pool *pgxpool.Pool, req models.CreatePaymentRequest, merchantID uuid.UUID, idempotencyKey uuid.UUID) (models.Payment, error) {
	req.Normalize()
	if err := validations.ValidateCreatePaymentRequest(req); err != nil {
		return models.Payment{}, err
	}
	_, err := repository.GetCustomerByID(ctx, pool, merchantID, req.CustomerID)
	if err != nil {
		if errors.Is(err, repository.ErrCustomerNotFound) {
			return models.Payment{}, err
		}
		return models.Payment{}, ErrInternalServerError
	}
	existentPayment, err := repository.GetPaymentByIdempotencyKey(ctx, pool, idempotencyKey, merchantID)
	if err == nil {
		return existentPayment, nil
	} else if !errors.Is(err, repository.ErrPaymentNotFound) {
		return models.Payment{}, ErrInternalServerError
	}
	payment := createPaymentFromCreatePaymentRequest(req, merchantID, idempotencyKey)
	createdPayment, err := repository.CreatePayment(ctx, pool, payment)
	if err != nil {
		return models.Payment{}, ErrInternalServerError
	}
	return createdPayment, nil
}

func ProcessPayment(ctx context.Context, pool *pgxpool.Pool, paymentID uuid.UUID, merchantID uuid.UUID) (models.Payment, error) {
	payment, err := repository.GetPaymentByID(ctx, pool, paymentID, merchantID)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return models.Payment{}, err
		}
		return models.Payment{}, ErrInternalServerError
	}
	if !validations.ValidatePaymentStatusTransition(payment.Status, models.PaymentStatusProcessing) {
		return models.Payment{}, ErrInvalidStatusTransition
	}
	updatedPayment, err := repository.UpdatePaymentStatus(ctx, pool, merchantID, paymentID, payment.Status, models.PaymentStatusProcessing)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return models.Payment{}, err
		}
		return models.Payment{}, ErrInternalServerError
	}
	return updatedPayment, nil
}
