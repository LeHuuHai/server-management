package es

import (
	"fmt"
	"strings"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/elastic/go-elasticsearch/v8"
)

func Connect(config *commonconfig.ElasticsearchConfig) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: strings.Split(config.URL, ","),
		// Username:  config.Username,
		// Password:  config.Password,
	}

	esclient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrConnectElasticsearch, err)
	}

	// Ping kiểm tra
	_, err = esclient.Ping()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrConnectElasticsearch, err)
	}

	return esclient, nil
}
