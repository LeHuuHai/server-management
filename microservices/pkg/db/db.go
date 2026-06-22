package db

import (
	"fmt"
	"log/slog"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(config *pkgconfig.PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		config.Host, config.Username, config.Password, config.Database, config.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect postgres error: %v", err)
	}

	slog.Info("Postgres connected", "host", config.Host, "port", config.Port, "database", config.Database, "user", config.Username)
	return db, nil
}
