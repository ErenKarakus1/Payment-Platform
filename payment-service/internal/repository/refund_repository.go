package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const getAllRefundsByPaymentIDQuery = `
	SELECT
		id,
		payment_id,
		merchant_id,
		amount_cents,
		created_at
	FROM refunds
	WHERE payment_id=$1
	AND merchant_id=$2
`

func GetAllRefundsByPaymentID(ctx context.Context, pool *pgxpool.Pool, merchantID uuid.UUID, paymentID uuid.UUID) ([]models.Refund, error) {
	var refunds []models.Refund
	rows, err := pool.Query(
		ctx,
		getAllRefundsByPaymentIDQuery,
		paymentID,
		merchantID,
	)
	if err != nil {
		return []models.Refund{}, errors.New("internal server error")
	}
	defer rows.Close()

	for rows.Next() {
		var refund models.Refund
		err := rows.Scan(
			&refund.ID,
			&refund.PaymentID,
			&refund.MerchantID,
			&refund.AmountCents,
			&refund.CreatedAt,
		)
		if err != nil {
			return []models.Refund{}, errors.New("internal server error")
		}
		refunds = append(refunds, refund)
	}
	if err := rows.Err(); err != nil {
		return []models.Refund{}, errors.New("internal server error")
	}
	return refunds, nil
}
