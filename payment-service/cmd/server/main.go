package main

import (
	"log"

	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/config"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/db"
	"github.com/ErenKarakus1/Payment-Platform/payment-service/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to Postgres!")

	router := gin.Default()

	router.POST("/customers", handlers.CreateCustomerHandler(pool))
	router.GET("/customers", handlers.GetAllCustomersHandler(pool))
	router.GET("/customers/:id", handlers.GetCustomerByIDHandler(pool))
	router.POST("/payments", handlers.CreatePaymentHandler(pool))
	router.GET("/payments", handlers.GetAllPaymentsHandler(pool))
	router.GET("/payments/:id", handlers.GetPaymentByIDHandler(pool))
	router.POST("/payments/:id/process", handlers.ProcessPaymentHandler(pool))
	router.POST("/payments/:id/refunds", handlers.RefundHandler(pool))
	router.GET("/payments/:id/refunds", handlers.GetAllRefundsByPaymentIDHandler(pool))

	/*
		REMOVE THIS PART IN PROD
			Payment provider will give success or fail result
			Status of a payment must change in process service instead of by these endpoints
	*/
	router.POST("/payments/:id/succeed", handlers.SucceedPaymentHandler(pool))
	router.POST("/payments/:id/fail", handlers.FailPaymentHandler(pool))

	router.Run(":8082")

}
