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

func GetPaymentByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		id := ctx.Param("id")
		if id == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "payment id is required"})
			return
		}
		parsedID, err := uuid.Parse(id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
			return
		}
		payment, err := repository.GetPaymentByID(ctx.Request.Context(), pool, parsedID, merchantID)
		if err != nil {
			if errors.Is(err, repository.ErrPaymentNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, payment)
	}
}

func GetAllPaymentsHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		payments, err := repository.GetAllPayments(ctx.Request.Context(), pool, merchantID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if payments == nil {
			payments = []models.Payment{}
		}
		ctx.JSON(http.StatusOK, payments)
	}
}

func ProcessPaymentHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		paymentID := ctx.Param("id")
		if paymentID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "payment id is required"})
			return
		}
		parsedPaymentID, err := uuid.Parse(paymentID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment id"})
			return
		}
		payment, err := repository.GetPaymentByID(ctx.Request.Context(), pool, parsedPaymentID, merchantID)
		if err != nil {
			if errors.Is(err, repository.ErrPaymentNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if !validations.ValidatePaymentStatusTransition(payment.Status, services.PaymentStatusProcessing) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "invalid status transition"})
			return
		}
		updatedPayment, err := repository.UpdatePaymentStatus(ctx.Request.Context(), pool, merchantID, parsedPaymentID, payment.Status, services.PaymentStatusProcessing)
		if err != nil {
			if errors.Is(err, repository.ErrPaymentNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, updatedPayment)
	}
}
