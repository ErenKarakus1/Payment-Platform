package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPaymentNotFound = errors.New("payment not found")

const createPaymentQuery = `
	INSERT INTO payments (
		id,
		merchant_id,
		customer_id,
		amount_cents,
		currency,
		status,
		idempotency_key
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7)
	RETURNING
		id,
		merchant_id,
		customer_id,
		amount_cents,
		refunded_amount_cents,
		currency,
		status,
		idempotency_key,
		created_at,
		updated_at
`
const getPaymentByIdempotencyKeyQuery = `
	SELECT
		id,
		merchant_id,
		customer_id,
		amount_cents,
		refunded_amount_cents,
		currency,
		status,
		idempotency_key,
		created_at,
		updated_at
	FROM payments
	WHERE idempotency_key=$1
	AND merchant_id=$2
`

func CreatePayment(ctx context.Context, pool *pgxpool.Pool, req models.Payment) (models.Payment, error) {
	var payment models.Payment
	err := pool.QueryRow(
		ctx,
		createPaymentQuery,
		req.ID,
		req.MerchantID,
		req.CustomerID,
		req.AmountCents,
		req.Currency,
		req.Status,
		req.IdempotencyKey,
	).Scan(
		&payment.ID,
		&payment.MerchantID,
		&payment.CustomerID,
		&payment.AmountCents,
		&payment.RefundedAmountCents,
		&payment.Currency,
		&payment.Status,
		&payment.IdempotencyKey,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		return models.Payment{}, errors.New("internal server error")
	}
	return payment, nil
}

func GetPaymentByIdempotencyKey(ctx context.Context, pool *pgxpool.Pool, idempotencyKey uuid.UUID, merchantID uuid.UUID) (models.Payment, error) {
	var payment models.Payment
	err := pool.QueryRow(
		ctx,
		getPaymentByIdempotencyKeyQuery,
		idempotencyKey,
		merchantID,
	).Scan(
		&payment.ID,
		&payment.MerchantID,
		&payment.CustomerID,
		&payment.AmountCents,
		&payment.RefundedAmountCents,
		&payment.Currency,
		&payment.Status,
		&payment.IdempotencyKey,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Payment{}, ErrPaymentNotFound
		}
		return models.Payment{}, errors.New("internal server error")
	}
	return payment, nil
}
