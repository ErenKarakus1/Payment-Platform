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

func CreateCustomerHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		customer := services.CreateCustomerFromCreateCustomerRequest(req, merchantID)
		createdCustomer, err := repository.CreateCustomer(ctx.Request.Context(), pool, customer)
		if err != nil {
			if errors.Is(err, repository.ErrEmailAlreadyUsed) {
				ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusCreated, createdCustomer)
	}
}

func GetAllCustomersHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		customers, err := repository.GetAllCustomers(ctx.Request.Context(), pool, merchantID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if customers == nil {
			customers = []models.Customer{}
		}
		ctx.JSON(http.StatusOK, customers)
	}
}

func GetCustomerByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		customerID := ctx.Param("id")
		if customerID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "customer id is required"})
			return
		}
		parsedCustomerID, err := uuid.Parse(customerID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
			return
		}
		customer, err := repository.GetCustomerByID(ctx.Request.Context(), pool, merchantID, parsedCustomerID)
		if err != nil {
			if errors.Is(err, repository.ErrCustomerNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, customer)
	}
}
