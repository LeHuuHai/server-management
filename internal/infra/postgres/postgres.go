package pg

import (
	"fmt"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(config *commonconfig.PostgresConfig) (*gorm.DB, error) {
	// đọc biến môi trường từ os
	host := config.Host
	user := config.Username
	password := config.Password
	dbname := config.Database
	port := config.Port

	// config của database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", host, user, password, dbname, port)

	// Mở kết nối tới database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrConnectPostgres, err)
	}

	return db, nil
}
