package main

import (
	"log"

	"github.com/ErenKarakus1/Payment-Platform/api-gateway/internal/config"
	"github.com/ErenKarakus1/Payment-Platform/api-gateway/internal/middleware"
	"github.com/ErenKarakus1/Payment-Platform/api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
)

func main() {

	cfg := config.LoadConfig()

	authProxy := proxy.NewProxy("http://localhost:8081")
	paymentProxy := proxy.NewProxy("http://localhost:8082")

	router := gin.Default()

	// Auth
	router.POST("/auth/register", authProxy)
	router.POST("/auth/login", authProxy)

	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// Customers
	protected.POST("/customers", paymentProxy)
	protected.GET("/customers", paymentProxy)
	protected.GET("/customers/:id", paymentProxy)

	// Payments
	protected.POST("/payments", middleware.RequireIdempotencyKey(), paymentProxy)
	protected.GET("/payments", paymentProxy)
	protected.GET("/payments/:id", paymentProxy)
	protected.POST("/payments/:id/process", paymentProxy)

	// Refunds
	protected.POST("/payments/:id/refunds", paymentProxy)
	protected.GET("/payments/:id/refunds", paymentProxy)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
