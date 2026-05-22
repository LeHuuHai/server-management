package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type MasterConfig struct {
	App   AppConfig
	DB    PostgresConfig
	Redis RedisConfig
	Kafka KafkaConfig
	ES    ElasticsearchConfig
}

type AppConfig struct {
	Port      int
	Host      string
	CyclePing int
}

type PostgresConfig struct {
	PgHost     string
	PgUsername string
	PgPassword string
	PgDatabase string
	PgPort     int
}

type RedisConfig struct {
	RedisURL      string
	RedisPassword string
	RedisDB       int
}

type KafkaConfig struct {
	KafkaBroker string
	//KafkaConsumerGroupId string
}

type ElasticsearchConfig struct {
	EsURL      string
	EsUsername string
	EsPassword string
}

func Load() (*MasterConfig, error) {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
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

	return &MasterConfig{
		App: AppConfig{
			Port:      appPort,
			Host:      os.Getenv("APP_HOST"),
			CyclePing: appCyclePing,
		},
		DB: PostgresConfig{
			PgHost:     os.Getenv("DB_HOST"),
			PgUsername: os.Getenv("DB_USER"),
			PgPassword: os.Getenv("DB_PASSWORD"),
			PgPort:     pgport,
			PgDatabase: os.Getenv("DB_DBNAME"),
		},
		Redis: RedisConfig{
			RedisURL:      os.Getenv("REDIS_URL"),
			RedisPassword: os.Getenv("REDIS_PASSWORD"),
			RedisDB:       redisdb,
		},
		Kafka: KafkaConfig{
			KafkaBroker: os.Getenv("KAFKA_BROKER"),
		},
		ES: ElasticsearchConfig{
			EsURL:      os.Getenv("ES_URL"),
			EsUsername: os.Getenv("ES_USER"),
			EsPassword: os.Getenv("ES_PASS"),
		},
	}, nil
}
