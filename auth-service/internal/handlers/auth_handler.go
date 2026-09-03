package handlers

import (
	"errors"
	"net/http"

	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/jwt"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/models"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/password"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/repository"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/services"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req models.RegisterRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		req.Normalize()
		if err := validation.ValidateRegisterRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user, err := services.CreateUserFromRegisterRequest(req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		createUserResponse, err := repository.CreateUser(ctx.Request.Context(), pool, user)
		if err != nil {
			if errors.Is(err, repository.ErrEmailAlreadyRegistered) {
				ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusCreated, createUserResponse)
	}
}

func LoginHandler(pool *pgxpool.Pool, jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req models.LoginRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		req.Normalize()
		if err := validation.ValidateLoginRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user, err := repository.GetUserByEmail(ctx.Request.Context(), pool, req.Email)
		if err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if err := password.CompareHashAndPassword(user.PasswordHash, req.Password); err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		token, err := jwt.GenerateTokenFromUser(jwtSecret, user.ID.String())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"token": token})
	}
}
