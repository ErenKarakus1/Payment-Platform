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

const (
	EventPaymentCreated           = "payment.created"
	EventPaymentProcessing        = "payment.processing"
	EventPaymentSucceeded         = "payment.succeeded"
	EventPaymentFailed            = "payment.failed"
	EventPaymentRefunded          = "payment.refunded"
	EventPaymentPartiallyRefunded = "payment.partially_refunded"
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
var ErrKafkaPublishEvent = errors.New("couldnt publish event to kafka")

func CreatePayment(ctx context.Context, pool *pgxpool.Pool, producer *kafka.Producer, req models.CreatePaymentRequest, merchantID uuid.UUID, idempotencyKey uuid.UUID) (models.Payment, error) {
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
	customer, err := repository.GetCustomerByID(ctx, pool, merchantID, createdPayment.CustomerID)
	if err != nil {
		return models.Payment{}, ErrInternalServerError
	}
	err = producer.PublishPaymentEvent(ctx, kafka.PaymentEvent{
		EventType:     EventPaymentCreated,
		PaymentID:     createdPayment.ID,
		MerchantID:    merchantID,
		CustomerID:    createdPayment.CustomerID,
		CustomerEmail: customer.Email,
		CustomerName:  customer.Name,
		AmountCents:   createdPayment.AmountCents,
		Currency:      createdPayment.Currency,
	})
	if err != nil {
		return models.Payment{}, ErrKafkaPublishEvent
	}
	return createdPayment, nil
}

func ProcessPayment(ctx context.Context, pool *pgxpool.Pool, producer *kafka.Producer, paymentID uuid.UUID, merchantID uuid.UUID) (models.Payment, error) {
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
	customer, err := repository.GetCustomerByID(ctx, pool, merchantID, updatedPayment.CustomerID)
	if err != nil {
		return models.Payment{}, ErrInternalServerError
	}
	err = producer.PublishPaymentEvent(ctx, kafka.PaymentEvent{
		EventType:     EventPaymentProcessing,
		PaymentID:     updatedPayment.ID,
		MerchantID:    merchantID,
		CustomerID:    updatedPayment.CustomerID,
		CustomerEmail: customer.Email,
		CustomerName:  customer.Name,
		AmountCents:   updatedPayment.AmountCents,
		Currency:      updatedPayment.Currency,
	})
	if err != nil {
		return models.Payment{}, ErrKafkaPublishEvent
	}
	return updatedPayment, nil
}

func SucceedPayment(ctx context.Context, pool *pgxpool.Pool, producer *kafka.Producer, paymentID uuid.UUID, merchantID uuid.UUID) (models.Payment, error) {
	payment, err := repository.GetPaymentByID(ctx, pool, paymentID, merchantID)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return models.Payment{}, err
		}
		return models.Payment{}, ErrInternalServerError
	}
	if !validations.ValidatePaymentStatusTransition(payment.Status, models.PaymentStatusSucceeded) {
		return models.Payment{}, ErrInvalidStatusTransition
	}
	updatedPayment, err := repository.UpdatePaymentStatus(ctx, pool, merchantID, paymentID, payment.Status, models.PaymentStatusSucceeded)
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
		EventType:     EventPaymentSucceeded,
		PaymentID:     updatedPayment.ID,
		MerchantID:    merchantID,
		CustomerID:    updatedPayment.CustomerID,
		CustomerEmail: customer.Email,
		CustomerName:  customer.Name,
		AmountCents:   updatedPayment.AmountCents,
		Currency:      updatedPayment.Currency,
	})
	if err != nil {
		return models.Payment{}, ErrKafkaPublishEvent
	}
	return updatedPayment, nil
}

func FailPayment(ctx context.Context, pool *pgxpool.Pool, producer *kafka.Producer, paymentID uuid.UUID, merchantID uuid.UUID) (models.Payment, error) {
	payment, err := repository.GetPaymentByID(ctx, pool, paymentID, merchantID)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return models.Payment{}, err
		}
		return models.Payment{}, ErrInternalServerError
	}
	if !validations.ValidatePaymentStatusTransition(payment.Status, models.PaymentStatusFailed) {
		return models.Payment{}, ErrInvalidStatusTransition
	}
	updatedPayment, err := repository.UpdatePaymentStatus(ctx, pool, merchantID, paymentID, payment.Status, models.PaymentStatusFailed)
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
		EventType:     EventPaymentFailed,
		PaymentID:     updatedPayment.ID,
		MerchantID:    merchantID,
		CustomerID:    updatedPayment.CustomerID,
		CustomerEmail: customer.Email,
		CustomerName:  customer.Name,
		AmountCents:   updatedPayment.AmountCents,
		Currency:      updatedPayment.Currency,
	})
	if err != nil {
		return models.Payment{}, ErrKafkaPublishEvent
	}
	return updatedPayment, nil
}
