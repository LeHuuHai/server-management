package db

import (
	"embed"
	"fmt"
	"log/slog"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func RunMigrations(config *pkgconfig.PostgresConfig, fs embed.FS, path string) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	sourceDriver, err := iofs.New(fs, path)
	if err != nil {
		return fmt.Errorf("failed to create source driver: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrate up: %v", err)
	}

	slog.Info("Database migrations completed successfully", "db", config.Database)
	return nil
}

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
