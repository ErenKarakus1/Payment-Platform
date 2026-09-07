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

var ErrInternalServerError = errors.New("internal server error")

func createCustomerFromCreateCustomerRequest(req models.CreateCustomerRequest, merchantID uuid.UUID) models.Customer {
	return models.Customer{
		ID:         uuid.New(),
		MerchantID: merchantID,
		Name:       req.Name,
		Email:      req.Email,
	}
}

func CreateCustomer(ctx context.Context, pool *pgxpool.Pool, merchantID uuid.UUID, req models.CreateCustomerRequest) (models.Customer, error) {
	req.Normalize()
	if err := validations.ValidateCreateCustomerRequest(req); err != nil {
		return models.Customer{}, err
	}
	customer := createCustomerFromCreateCustomerRequest(req, merchantID)
	createdCustomer, err := repository.CreateCustomer(ctx, pool, customer)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyUsed) {
			return models.Customer{}, repository.ErrEmailAlreadyUsed
		}
		return models.Customer{}, ErrInternalServerError
	}
	return createdCustomer, nil
}
