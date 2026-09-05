package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailAlreadyUsed = errors.New("email is already used")

const createCustomerQuery = `
	INSERT INTO customers (
		id,
		merchant_id,
		name,
		email
	)
	VALUES ($1,$2,$3,$4)
	RETURNING
		id,
		merchant_id,
		name,
		email,
		created_at
`

func CreateCustomer(ctx context.Context, pool *pgxpool.Pool, customer models.Customer) (models.Customer, error) {
	var createdCustomer models.Customer
	err := pool.QueryRow(
		ctx,
		createCustomerQuery,
		customer.ID,
		customer.MerchantID,
		customer.Name,
		customer.Email,
	).Scan(
		&createdCustomer.ID,
		&createdCustomer.MerchantID,
		&createdCustomer.Name,
		&createdCustomer.Email,
		&createdCustomer.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.Customer{}, ErrEmailAlreadyUsed
		}
		return models.Customer{}, errors.New("internal server error")
	}
	return createdCustomer, nil
}
