package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/google/uuid"
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

const getAllCustomersQuery = `
	SELECT
		id,
		merchant_id,
		name,
		email,
		created_at
	FROM customers
	WHERE merchant_id=$1
	ORDER BY created_at DESC
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

func GetAllCustomers(ctx context.Context, pool *pgxpool.Pool, merchantID uuid.UUID) ([]models.Customer, error) {

	rows, err := pool.Query(
		ctx,
		getAllCustomersQuery,
		merchantID,
	)
	if err != nil {
		return []models.Customer{}, errors.New("internal server error")
	}
	defer rows.Close()

	var customers []models.Customer

	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(
			&customer.ID,
			&customer.MerchantID,
			&customer.Name,
			&customer.Email,
			&customer.CreatedAt,
		)
		if err != nil {
			return []models.Customer{}, errors.New("internal server error")
		}
		customers = append(customers, customer)
	}
	if err := rows.Err(); err != nil {
		return []models.Customer{}, errors.New("internal server error")
	}
	return customers, nil
}
