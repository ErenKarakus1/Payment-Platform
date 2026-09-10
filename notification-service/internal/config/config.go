package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return Config{
		SMTPHost:     getEnv("SMTP_HOST"),
		SMTPPort:     getEnv("SMTP_PORT"),
		SMTPUsername: getEnv("SMTP_USERNAME"),
		SMTPPassword: getEnv("SMTP_PASSWORD"),
	}
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s is not set", key)
	}
	return value
}
