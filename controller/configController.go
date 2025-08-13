package controller

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress string
	ServerPort    string
}

var Cfg *Config

func LoadConfig() {
	_ = godotenv.Load()

	Cfg = &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", "localhost"),
		ServerPort:    getEnv("SERVER_PORT", "1759"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
