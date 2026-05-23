package commonconfig

type PostgresConfig struct {
	Host     string
	Username string
	Password string
	Database string
	Port     int
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

type KafkaWriterConfig struct {
	Broker string
}

type ElasticsearchConfig struct {
	URL      string
	Username string
	Password string
}
