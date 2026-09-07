package services

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/jwt"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/models"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/password"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/repository"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/validation"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInternalServerError = errors.New("internal server error")
var ErrUnauthorized = errors.New("invalid email or password")

func createUserFromRegisterRequest(r models.RegisterRequest) (models.User, error) {
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

func Register(ctx context.Context, pool *pgxpool.Pool, req models.RegisterRequest) (models.CreateUserResponse, error) {
	req.Normalize()
	if err := validation.ValidateRegisterRequest(req); err != nil {
		return models.CreateUserResponse{}, err
	}
	user, err := createUserFromRegisterRequest(req)
	if err != nil {
		return models.CreateUserResponse{}, ErrInternalServerError
	}
	createUserResponse, err := repository.CreateUser(ctx, pool, user)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyRegistered) {
			return models.CreateUserResponse{}, err
		}
		return models.CreateUserResponse{}, ErrInternalServerError
	}
	return createUserResponse, nil
}

func Login(ctx context.Context, pool *pgxpool.Pool, jwtSecret string, req models.LoginRequest) (string, error) {
	req.Normalize()
	if err := validation.ValidateLoginRequest(req); err != nil {
		return "", err
	}
	user, err := repository.GetUserByEmail(ctx, pool, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrUnauthorized
		}
		return "", ErrInternalServerError
	}
	if err := password.CompareHashAndPassword(user.PasswordHash, req.Password); err != nil {
		return "", ErrUnauthorized
	}
	token, err := jwt.GenerateTokenFromUser(jwtSecret, user.ID.String())
	if err != nil {
		return "", ErrInternalServerError
	}
	return token, nil
}
