package es

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	pkgconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
	"github.com/elastic/go-elasticsearch/v8"
)

func Connect(config *pkgconfig.ElasticsearchConfig) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: strings.Split(config.URL, ","),
		Username:  config.Username,
		Password:  config.Password,
	}

	esclient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrConnectElasticsearch, err)
	}

	// Ping check
	_, err = esclient.Ping()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrConnectElasticsearch, err)
	}
	slog.Info("Elasticsearch connected", "urls", cfg.Addresses)
	return esclient, nil
}
