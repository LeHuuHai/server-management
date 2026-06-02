package agentconfig

import (
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

func Load() (*Config, error) {
	err := godotenv.Load("./config/agent/.env.agent")
	if err != nil {
		panic("Error loading .env file")
	}

	cycleHeartbeat, err := strconv.Atoi(os.Getenv("APP_CYCLE_HEARBEAT"))
	if err != nil {
		return nil, err
	}

	return &Config{
		AppConfig: &AppConfig{
			ServerID:        os.Getenv("APP_SERVER_ID"),
			HeartbeatURL:    os.Getenv("APP_HEARTBEAT_URL"),
			HeartbeatAPIKey: os.Getenv("APP_HEARTBEAT_KEY"),
			CycleHeartbeat:  cycleHeartbeat,
		},
	}, nil
}
