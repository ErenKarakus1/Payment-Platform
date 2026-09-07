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

const getPaymentByIDQuery = `
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
	WHERE id=$1
	AND merchant_id=$2
`

const getAllPaymentsQuery = `
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
	WHERE merchant_id=$1
`

const updatePaymentStatusQuery = `
	UPDATE payments
	SET
		status=$1,
		updated_at=NOW()
	WHERE id=$2
	AND merchant_id=$3
	AND status=$4
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

func GetPaymentByID(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, merchantID uuid.UUID) (models.Payment, error) {
	var payment models.Payment
	err := pool.QueryRow(
		ctx,
		getPaymentByIDQuery,
		id,
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

func GetAllPayments(ctx context.Context, pool *pgxpool.Pool, merchantID uuid.UUID) ([]models.Payment, error) {
	rows, err := pool.Query(
		ctx,
		getAllPaymentsQuery,
		merchantID,
	)
	if err != nil {
		return []models.Payment{}, errors.New("internal server error")
	}
	defer rows.Close()
	var payments []models.Payment
	for rows.Next() {
		var payment models.Payment
		err := rows.Scan(
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
			return []models.Payment{}, errors.New("internal server error")
		}
		payments = append(payments, payment)
	}
	if err := rows.Err(); err != nil {
		return []models.Payment{}, errors.New("internal server error")
	}
	return payments, nil
}

func UpdatePaymentStatus(ctx context.Context, pool *pgxpool.Pool, merchantID uuid.UUID, paymentID uuid.UUID, currentStatus string, targetStatus string) (models.Payment, error) {
	var payment models.Payment
	err := pool.QueryRow(
		ctx,
		updatePaymentStatusQuery,
		targetStatus,
		paymentID,
		merchantID,
		currentStatus,
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
