package config

import (
	"os"
	"strconv"
	"strings"

	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
)

type AppConfig struct {
	Port        int
	Host        string
	CORSOrigins []string
}

type Config struct {
	AppConfig   *AppConfig
	JWTConfig   *pkgconfig.JWTConfig
	DBConfig    *pkgconfig.PostgresConfig
	RedisConfig *pkgconfig.RedisConfig
}

func Load() (*Config, error) {
	accessExpired, err := strconv.Atoi(os.Getenv("JWT_ACCESS_EXPIRED"))
	if err != nil {
		return nil, err
	}

	refreshExpired, err := strconv.Atoi(os.Getenv("JWT_REFRESH_EXPIRED"))
	if err != nil {
		return nil, err
	}

	pgport, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, err
	}

	redisdb, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		return nil, err
	}

	appPort, err := strconv.Atoi(os.Getenv("APP_PORT"))
	if err != nil {
		return nil, err
	}

	cfg := Config{
		AppConfig: &AppConfig{
			Port:        appPort,
			Host:        os.Getenv("APP_HOST"),
			CORSOrigins: strings.Split(os.Getenv("APP_CORS_ORIGIN"), ","),
		},
		JWTConfig: &pkgconfig.JWTConfig{
			AccessSecret:   pkgconfig.ReadSecret("jwt_access_secret"),
			RefreshSecret:  pkgconfig.ReadSecret("jwt_refresh_secret"),
			AccessExpired:  accessExpired,
			RefreshExpired: refreshExpired,
		},
		DBConfig: &pkgconfig.PostgresConfig{
			Host:     os.Getenv("DB_HOST"),
			Username: os.Getenv("DB_USER"),
			Password: pkgconfig.ReadSecret("auth_db_password"),
			Port:     pgport,
			Database: os.Getenv("DB_DBNAME"),
		},
		RedisConfig: &pkgconfig.RedisConfig{
			URL: os.Getenv("REDIS_URL"),
			DB:  redisdb,
		},
	}
	return &cfg, nil
}
