package services

import (
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/google/uuid"
)

func CreateCustomerFromCreateCustomerRequest(req models.CreateCustomerRequest, merchantID uuid.UUID) models.Customer {
	return models.Customer{
		ID:         uuid.New(),
		MerchantID: merchantID,
		Name:       req.Name,
		Email:      req.Email,
	}
}
