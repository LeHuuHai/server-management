#!/bin/bash

set -e

ES=http://elasticsearch:9200

echo "Waiting Elasticsearch..."

until curl -s $ES >/dev/null; do
  sleep 2
done

echo "Elasticsearch is ready"

# =========================
# ILM POLICY
# =========================
echo "Creating ILM policy: heartbeat-policy"

curl -s -o /dev/null -w "%{http_code}" -X PUT "$ES/_ilm/policy/heartbeat-policy" \
  -H "Content-Type: application/json" \
  -d '{
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
  }'

# =========================
# COMPONENT TEMPLATE
# =========================
echo "Creating component template: heartbeat-base"

curl -s -o /dev/null -w "%{http_code}" -X PUT "$ES/_component_template/heartbeat-base" \
  -H "Content-Type: application/json" \
  -d '{
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
  }'

# =========================
# INDEX TEMPLATE
# =========================
echo "Creating index template: heartbeat-template"

curl -s -o /dev/null -w "%{http_code}" -X PUT "$ES/_index_template/heartbeat-template" \
  -H "Content-Type: application/json" \
  -d '{
    "index_patterns": ["heartbeat-*"],
    "template": {
      "settings": {
        "index.lifecycle.name": "heartbeat-policy",
        "index.lifecycle.rollover_alias": "heartbeat",
        "number_of_shards": 1
      }
    },
    "composed_of": ["heartbeat-base"]
  }'

# =========================
# INITIAL INDEX
# =========================
echo "Creating initial index: heartbeat-000001"

HTTP=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$ES/heartbeat-000001" \
  -H "Content-Type: application/json" \
  -d '{
    "aliases": {
      "heartbeat": {
        "is_write_index": true
      }
    }
  }')

if [ "$HTTP" = "200" ] || [ "$HTTP" = "201" ]; then
  echo "Index created or already exists"
else
  echo "Index creation returned HTTP $HTTP (maybe already exists)"
fi

echo "Done"