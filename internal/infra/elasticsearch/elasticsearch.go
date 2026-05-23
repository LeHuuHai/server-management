package es

import (
	"strings"

	commonconfig "github.com/LeHuuHai/server-management/config/common"
	"github.com/elastic/go-elasticsearch/v8"
)

func Connect(config *commonconfig.ElasticsearchConfig) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: strings.Split(config.URL, ","),
		Username:  config.Username,
		Password:  config.Password,
	}

	esclient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	// Ping kiểm tra
	_, err = esclient.Ping()
	if err != nil {
		return nil, err
	}

	return esclient, nil
}
