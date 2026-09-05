package validations

import (
	"errors"

	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/google/uuid"
)

var supportedCurrencies = map[string]struct{}{
	"USD": {},
	"TRY": {},
	"EUR": {},
}

func ValidateCreatePaymentRequest(r models.CreatePaymentRequest) error {
	if r.AmountCents == 0 {
		return errors.New("amount cant be zero")
	}
	if r.AmountCents < 0 {
		return errors.New("amount cant be negative")
	}
	if r.Currency == "" {
		return errors.New("currency is required")
	}
	if _, ok := supportedCurrencies[r.Currency]; !ok {
		return errors.New("unsupported currency")
	}
	if r.CustomerID == uuid.Nil {
		return errors.New("customer id is required")
	}
	return nil
}
