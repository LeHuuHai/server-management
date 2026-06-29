#!/bin/bash
set -e

declare -A secrets=(
    ["auth_db_password"]="password"
    ["jwt_access_secret"]="your_access_secret"
    ["jwt_refresh_secret"]="your_refresh_secret"
    ["server_db_password"]="password"
    ["monitor_db_password"]="password"
    ["download_report_api_key"]="download_report_api_key"
    ["heartbeat_api_key"]="heartbeat_api_key"
    ["gomail_password"]="smtp_password"
)

echo "Initializing Docker Swarm secrets..."

for secret in "${!secrets[@]}"; do
    if ! docker secret ls --format '{{.Name}}' | grep -q "^${secret}$"; then
        echo -n "${secrets[$secret]}" | docker secret create "$secret" -
        echo "Created secret: $secret"
    else
        echo "Secret already exists: $secret"
    fi
done

echo "Secrets initialized successfully."
