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

type KafkaReaderConfig struct {
	Broker          string
	Topic           string
	ConsumerGroupId string
}

type ElasticsearchConfig struct {
	URL      string
	Username string
	Password string
}

type GomailConfig struct {
	Addr     string
	Port     int
	From     string
	Password string
}
