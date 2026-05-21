package es

import (
	"strings"

	"github.com/LeHuuHai/server-management/config"
	"github.com/elastic/go-elasticsearch/v8"
)

func Connect(config *config.MasterConfig) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: strings.Split(config.ES.EsURL, ","),
		Username:  config.ES.EsUsername,
		Password:  config.ES.EsPassword,
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
