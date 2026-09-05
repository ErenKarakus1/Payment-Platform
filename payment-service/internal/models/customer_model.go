package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID         uuid.UUID `json:"id"`
	MerchantID uuid.UUID `json:"merchant_id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (r *CreateCustomerRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
}
