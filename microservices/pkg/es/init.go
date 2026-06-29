package es

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

func InitHeartbeatIndex(client *elasticsearch.Client) error {
	// 1. ILM POLICY
	ilmPolicy := `{
		"policy": {
			"phases": {
				"hot": {
					"min_age": "0ms",
					"actions": {
						"rollover": {
							"max_age": "1h"
						}
					}
				},
				"delete": {
					"min_age": "3h",
					"actions": {
						"delete": {}
					}
				}
			}
		}
	}`
	res, err := client.ILM.PutLifecycle(
		"heartbeat-policy",
		client.ILM.PutLifecycle.WithBody(strings.NewReader(ilmPolicy)),
	)
	if err != nil {
		return fmt.Errorf("failed to create ILM policy: %v", err)
	}
	res.Body.Close()

	// 2. COMPONENT TEMPLATE
	componentTemplate := `{
		"template": {
			"settings": {
				"number_of_shards": 1
			},
			"mappings": {
				"dynamic": false,
				"properties": {
					"server_id": {
						"type": "keyword"
					},
					"status": {
						"type": "keyword"
					},
					"timestamp": {
						"type": "date"
					}
				}
			}
		}
	}`
	res, err = client.Cluster.PutComponentTemplate(
		"heartbeat-base",
		strings.NewReader(componentTemplate),
	)
	if err != nil {
		return fmt.Errorf("failed to create component template: %v", err)
	}
	res.Body.Close()

	// 3. INDEX TEMPLATE
	indexTemplate := `{
		"index_patterns": ["heartbeat-*"],
		"template": {
			"settings": {
				"index.lifecycle.name": "heartbeat-policy",
				"index.lifecycle.rollover_alias": "heartbeat",
				"number_of_shards": 1
			}
		},
		"composed_of": ["heartbeat-base"]
	}`
	res, err = client.Indices.PutIndexTemplate(
		"heartbeat-template",
		strings.NewReader(indexTemplate),
	)
	if err != nil {
		return fmt.Errorf("failed to create index template: %v", err)
	}
	res.Body.Close()

	// 4. INITIAL INDEX (Ignore error if it already exists)
	initialIndex := `{
		"aliases": {
			"heartbeat": {
				"is_write_index": true
			}
		}
	}`
	res, err = client.Indices.Create(
		"heartbeat-000001",
		client.Indices.Create.WithBody(strings.NewReader(initialIndex)),
	)
	if err == nil {
		res.Body.Close()
	}

	slog.Info("Elasticsearch heartbeat index and policies initialized successfully")
	return nil
}
