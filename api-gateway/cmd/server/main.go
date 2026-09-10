package main

import (
	"log"
	"time"

	"github.com/ErenKarakus1/Payment-Platform/api-gateway/internal/config"
	"github.com/ErenKarakus1/Payment-Platform/api-gateway/internal/middlewares"
	"github.com/ErenKarakus1/Payment-Platform/api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {

	cfg := config.LoadConfig()

	authProxy := proxy.NewProxy("http://localhost:8081")
	paymentProxy := proxy.NewProxy("http://localhost:8082")

	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer redisClient.Close()

	router := gin.Default()

	// Auth
	router.POST("/auth/register", authProxy)
	router.POST("/auth/login", authProxy)

	protected := router.Group("/")
	protected.Use(middlewares.AuthMiddleware(cfg.JWTSecret))

	// Customers
	protected.POST("/customers", middlewares.RateLimiter(redisClient, 20, time.Minute), paymentProxy)
	protected.GET("/customers", middlewares.RateLimiter(redisClient, 60, time.Minute), paymentProxy)
	protected.GET("/customers/:id", middlewares.RateLimiter(redisClient, 60, time.Minute), paymentProxy)

	// Payments
	protected.POST("/payments", middlewares.RequireIdempotencyKey(), middlewares.RateLimiter(redisClient, 20, time.Minute), paymentProxy)
	protected.GET("/payments", middlewares.RateLimiter(redisClient, 60, time.Minute), paymentProxy)
	protected.GET("/payments/:id", middlewares.RateLimiter(redisClient, 60, time.Minute), paymentProxy)
	protected.POST("/payments/:id/process", middlewares.RateLimiter(redisClient, 20, time.Minute), paymentProxy)

	// Refunds
	protected.POST("/payments/:id/refunds", middlewares.RateLimiter(redisClient, 10, time.Minute), paymentProxy)
	protected.GET("/payments/:id/refunds", middlewares.RateLimiter(redisClient, 60, time.Minute), paymentProxy)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
