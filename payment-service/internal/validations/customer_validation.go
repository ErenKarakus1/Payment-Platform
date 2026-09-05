package validations

import (
	"errors"
	"net/mail"

	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
)

func ValidateCreateCustomerRequest(r models.CreateCustomerRequest) error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if len(r.Name) < 3 {
		return errors.New("name must be at least 3 characters")
	}
	if len(r.Name) > 50 {
		return errors.New("name must be at most 50 characters")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return errors.New("invalid email")
	}
	return nil
}
