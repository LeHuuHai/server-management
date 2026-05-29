package masterconfig

import (
	"net/mail"
	"os"
	"strconv"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	"github.com/joho/godotenv"
)

type Config struct {
	AppConfig   *AppConfig
	JWTConfig   *JWTConfig
	DBConfig    *commonconfig.PostgresConfig
	RedisConfig *commonconfig.RedisConfig
	KafkaConfig *commonconfig.KafkaConfig
	ESConfig    *commonconfig.ElasticsearchConfig
}

type AppConfig struct {
	Port      int
	Host      string
	CyclePing int
	AdMail    string
}

type JWTConfig struct {
	AccessSecret   string
	RefreshSecret  string
	AccessExpired  int
	RefreshExpired int
}

func Load() (*Config, error) {
	err := godotenv.Load("./config/master/.env.master")
	if err != nil {
		panic("Error loading .env file")
	}

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

	appCyclePing, err := strconv.Atoi(os.Getenv("APP_CYCLE_PING"))
	if err != nil {
		return nil, err
	}

	_, err = mail.ParseAddress(os.Getenv("APP_ADMAIL"))
	if err != nil {
		return nil, err
	}

	return &Config{
		AppConfig: &AppConfig{
			Port:      appPort,
			Host:      os.Getenv("APP_HOST"),
			CyclePing: appCyclePing,
			AdMail:    os.Getenv("APP_ADMAIL"),
		},
		JWTConfig: &JWTConfig{
			AccessSecret:   os.Getenv("JWT_ACCESS_SECRET"),
			RefreshSecret:  os.Getenv("JWT_REFRESH_SECRET"),
			AccessExpired:  accessExpired,
			RefreshExpired: refreshExpired,
		},
		DBConfig: &commonconfig.PostgresConfig{
			Host:     os.Getenv("DB_HOST"),
			Username: os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Port:     pgport,
			Database: os.Getenv("DB_DBNAME"),
		},
		RedisConfig: &commonconfig.RedisConfig{
			URL:      os.Getenv("REDIS_URL"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisdb,
		},
		KafkaConfig: &commonconfig.KafkaConfig{
			Writer: &commonconfig.KafkaWriterConfig{
				Broker: os.Getenv("KAFKA_BROKER"),
			},
			Topics: map[string]string{
				"ping": os.Getenv("KAFKA_PING_TOPIC"),
				"mail": os.Getenv("KAFKA_MAIL_TOPIC"),
			},
		},
		ESConfig: &commonconfig.ElasticsearchConfig{
			URL:      os.Getenv("ES_URL"),
			Username: os.Getenv("ES_USER"),
			Password: os.Getenv("ES_PASSWORD"),
			Index:    os.Getenv("ES_INDEX"),
		},
	}, nil
}
