package handlers

import (
	"errors"
	"net/http"

	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/models"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/repository"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/services"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetAllRefundsByPaymentIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		refunds, err := services.GetAllRefundsByPaymentID(ctx.Request.Context(), pool, parsedPaymentID, merchantID)
		if err != nil {
			if errors.Is(err, repository.ErrPaymentNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
				return
			} else if errors.Is(err, services.ErrInternalServerError) {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if refunds == nil {
			refunds = []models.Refund{}
		}
		ctx.JSON(http.StatusOK, refunds)
	}
}
