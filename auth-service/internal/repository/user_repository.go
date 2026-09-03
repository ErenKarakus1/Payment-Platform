package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailAlreadyRegistered = errors.New("email is already registered")
var ErrUserNotFound = errors.New("user not found")

const createUserQuery = `
	INSERT INTO users (
		id,
		name,
		email,
		password_hash
	)
	VALUES ($1,$2,$3,$4)
	RETURNING
		id,
		name,
		email,
		created_at,
		updated_at
`

const getUserByEmailQuery = `
	SELECT
		id,
		name,
		email,
		password_hash,
		created_at,
		updated_at
	FROM users
	WHERE email=$1
`

func CreateUser(ctx context.Context, pool *pgxpool.Pool, user models.User) (models.CreateUserResponse, error) {
	var createUserResponse models.CreateUserResponse
	err := pool.QueryRow(
		ctx,
		createUserQuery,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
	).Scan(
		&createUserResponse.ID,
		&createUserResponse.Name,
		&createUserResponse.Email,
		&createUserResponse.CreatedAt,
		&createUserResponse.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.CreateUserResponse{}, ErrEmailAlreadyRegistered
		}
		return models.CreateUserResponse{}, errors.New("internal server error")
	}
	return createUserResponse, nil
}

func GetUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (models.User, error) {
	var user models.User
	err := pool.QueryRow(
		ctx,
		getUserByEmailQuery,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, errors.New("internal server error")
	}
	return user, nil
}
