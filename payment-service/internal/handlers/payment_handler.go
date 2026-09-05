package handlers

import (
	"errors"
	"net/http"

	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/repository"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/services"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/utils"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/validations"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePaymentHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		merchantID, err := utils.GetMerchantID(ctx)
		if err != nil {
			if errors.Is(err, utils.ErrMissingMerchantID) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "merchant id is required"})
				return
			} else if errors.Is(err, utils.ErrInvalidMerchantID) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid merchant id"})
				return
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}
		idempotencyKey := ctx.GetHeader("Idempotency-Key")
		if idempotencyKey == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "idempotency key is required"})
			return
		}
		parsedIdempotencyKey, err := uuid.Parse(idempotencyKey)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid idempotency key"})
			return
		}
		existentPayment, err := repository.GetPaymentByIdempotencyKey(ctx.Request.Context(), pool, parsedIdempotencyKey, merchantID)
		if err == nil {
			ctx.JSON(http.StatusCreated, existentPayment)
			return
		} else if !errors.Is(err, repository.ErrPaymentNotFound) {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		var req models.CreatePaymentRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		req.Normalize()
		if err := validations.ValidateCreatePaymentRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err = repository.GetCustomerByID(ctx.Request.Context(), pool, merchantID, req.CustomerID)
		if err != nil {
			if errors.Is(err, repository.ErrCustomerNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		payment := services.CreatePaymentFromCreatePaymentRequest(req, merchantID, parsedIdempotencyKey)
		createdPayment, err := repository.CreatePayment(ctx.Request.Context(), pool, payment)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusCreated, createdPayment)
	}
}
