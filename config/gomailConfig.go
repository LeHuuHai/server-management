package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type GomailConfig struct {
	Addr     string
	Port     int
	From     string
	Password string
}

func LoadGomailConfig() (*GomailConfig, error) {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	port, err := strconv.Atoi(os.Getenv("GOMAIL_PORT"))
	if err != nil {
		return nil, err
	}

	return &GomailConfig{
		Addr:     os.Getenv("GOMAIL_ADDR"),
		Port:     port,
		From:     os.Getenv("GOMAIL_FROM"),
		Password: os.Getenv("GOMAIL_PASSWORD"),
	}, nil
}
