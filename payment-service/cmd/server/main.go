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
	router.Run(":8082")

}
