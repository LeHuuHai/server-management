package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppConfig *AppConfig
}

type AppConfig struct {
	ServerID        string
	HeartbeatURL    string
	HeartbeatAPIKey string
	CycleHeartbeat  int
}

func (c *AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("server_id", c.ServerID),
		slog.Any("heartbeat_url", c.HeartbeatURL),
		slog.Any("heartbeat_api_key", c.HeartbeatAPIKey),
		slog.Any("cycle_heartbeat", c.CycleHeartbeat),
	)
}

func (c *Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("app", c.AppConfig),
	)
}

func Load() (*Config, error) {
	// Try local path first (for running from service root)
	err := godotenv.Load(".env")
	if err != nil {
		// Fallback to workspace path
		err = godotenv.Load("./microservices/agent/.env")
		if err != nil {
			slog.Warn("Error loading .env file, fallback to OS env vars")
		}
	}

	cycleHeartbeat, err := strconv.Atoi(os.Getenv("APP_CYCLE_HEARTBEAT"))
	if err != nil {
		cycleHeartbeat = 5000 // Default 5s
	}

	cfg := Config{
		AppConfig: &AppConfig{
			ServerID:        os.Getenv("APP_SERVER_ID"),
			HeartbeatURL:    os.Getenv("APP_HEARTBEAT_URL"),
			HeartbeatAPIKey: os.Getenv("APP_HEARTBEAT_KEY"),
			CycleHeartbeat:  cycleHeartbeat,
		},
	}
	slog.Info("Config loaded", "config", &cfg)
	return &cfg, nil
}
