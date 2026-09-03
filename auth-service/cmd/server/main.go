package main

import (
	"log"

	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/config"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/db"
	"github.com/ErenKarakus1/Payment-Platform/auth-service/internal/handlers"
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

	router.POST("/auth/register", handlers.RegisterHandler(pool))
	router.POST("/auth/login", handlers.LoginHandler(pool, cfg.JWTSecret))
	if err := router.Run(":8081"); err != nil {
		log.Fatal(err)
	}

}
