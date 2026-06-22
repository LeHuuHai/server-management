package config

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

type JWTConfig struct {
	AccessSecret   string
	RefreshSecret  string
	AccessExpired  int
	RefreshExpired int
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

func (c *JWTConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("access_secret", c.AccessSecret),
		slog.String("refresh_secret", c.RefreshSecret),
		slog.Int("access_expired", c.AccessExpired),
		slog.Int("refresh_expired", c.RefreshExpired),
	)
}
