package database

import (
	"fmt"

	"github.com/LeHuuHai/server-management/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(config *config.Config) (*gorm.DB, error) {
	// đọc biến môi trường từ os
	host := config.DB.PgHost
	user := config.DB.PgUsername
	password := config.DB.PgPassword
	dbname := config.DB.PgDatabase
	port := config.DB.PgPort

	// config của database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", host, user, password, dbname, port)

	// Mở kết nối tới database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
