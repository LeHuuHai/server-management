package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DB    PostgresConfig
	Redis RedisConfig
	Kafka KafkaConfig
	ES    ElasticsearchConfig
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
	KafkaTopic string
}

type ElasticsearchConfig struct {
	EsURL      string
	EsUsername string
	EsPassword string
}

func Load() (*Config, error) {
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

	return &Config{
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
			KafkaTopic:  os.Getenv("KAFKA_TOPIC"),
		},
		ES: ElasticsearchConfig{
			EsURL:      os.Getenv("ES_URL"),
			EsUsername: os.Getenv("ES_USER"),
			EsPassword: os.Getenv("ES_PASS"),
		},
	}, nil
}
