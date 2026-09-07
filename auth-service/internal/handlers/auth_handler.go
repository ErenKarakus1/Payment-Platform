package handlers

import (
	"errors"
	"net/http"

	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/models"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/repository"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/services"
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
		createUserResponse, err := services.Register(ctx.Request.Context(), pool, req)
		if err != nil {
			if errors.Is(err, services.ErrInternalServerError) {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			} else if errors.Is(err, repository.ErrEmailAlreadyRegistered) {
				ctx.JSON(http.StatusConflict, gin.H{"error": "email is already registered"})
				return
			}
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, createUserResponse)
	}
}

func LoginHandler(pool *pgxpool.Pool, jwtSecret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req models.LoginRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		token, err := services.Login(ctx.Request.Context(), pool, jwtSecret, req)
		if err != nil {
			if errors.Is(err, services.ErrInternalServerError) {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			} else if errors.Is(err, services.ErrUnauthorized) {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
				return
			}
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"token": token})
	}
}
