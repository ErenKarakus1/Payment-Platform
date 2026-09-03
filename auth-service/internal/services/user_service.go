package services

import (
	"errors"

	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/models"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/password"
	"github.com/google/uuid"
)

func CreateUserFromRegisterRequest(r models.RegisterRequest) (models.User, error) {
	passwordHash, err := password.GeneratePasswordHash(r.Password)
	if err != nil {
		return models.User{}, errors.New("couldnt create user from register request")
	}
	return models.User{
		ID:           uuid.New(),
		Name:         r.Name,
		Email:        r.Email,
		PasswordHash: passwordHash,
	}, nil
}
