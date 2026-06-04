package commonconfig

import "log/slog"

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
	ConsumerGroupId string
}

type KafkaConfig struct {
	Writer *KafkaWriterConfig
	Reader *KafkaReaderConfig
	Topics map[string]string
}

type ElasticsearchConfig struct {
	URL      string
	Username string
	Password string
	Index    string
}

func (c *PostgresConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("host", c.Host),
		slog.Any("username", c.Username),
		slog.Any("password", c.Password),
		slog.Any("database", c.Database),
		slog.Any("port", c.Port),
	)
}

func (c *RedisConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("url", c.URL),
		slog.Any("password", c.Password),
		slog.Any("db", c.DB),
	)
}

func (c *KafkaWriterConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("broker", c.Broker),
	)
}

func (c *KafkaReaderConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("broker", c.Broker),
		slog.Any("consumer_group_id", c.ConsumerGroupId),
	)
}

func (c *KafkaConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("writer", c.Writer),
		slog.Any("reader", c.Reader),
		slog.Any("topics", c.Topics),
	)
}

func (c *ElasticsearchConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("url", c.URL),
		slog.Any("username", c.Username),
		slog.Any("password", c.Password),
		slog.Any("index", c.Index),
	)
}
