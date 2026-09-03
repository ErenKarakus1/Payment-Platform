package validation

import (
	"errors"
	"net/mail"

	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/models"
)

func ValidateRegisterRequest(r models.RegisterRequest) error {
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

	if r.Password == "" {
		return errors.New("password is required")
	}

	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	if len(r.Password) > 128 {
		return errors.New("password must be at most 128 characters")
	}
	return nil
}
