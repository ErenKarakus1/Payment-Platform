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
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateCustomerHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, err := utils.GetUserID(ctx)
		if err != nil {
			if errors.Is(err, utils.ErrMissingUserID) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
				return
			} else if errors.Is(err, utils.ErrInvalidUserID) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
				return
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
		}
		var req models.CreateCustomerRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		req.Normalize()
		if err := validations.ValidateCreateCustomerRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		customer := services.CreateCustomerFromCreateCustomerRequest(req, userID)
		createdCustomer, err := repository.CreateCustomer(ctx.Request.Context(), pool, customer)
		if err != nil {
			if errors.Is(err, repository.ErrEmailAlreadyUsed) {
				ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, createdCustomer)
	}
}
