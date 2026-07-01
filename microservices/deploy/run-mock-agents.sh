#!/bin/bash
set -e

# Configuration
NETWORK_NAME="sm_network"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-hariii31}"
IMAGE_NAME="${DOCKER_REGISTRY}/server-management-agent:v1.0.0"
HEARTBEAT_URL="http://heartbeat-gateway:8080/heartbeat"
API_KEY="heartbeat_api_key"

echo "Checking if overlay network '$NETWORK_NAME' exists..."
if ! docker network inspect "$NETWORK_NAME" >/dev/null 2>&1; then
  echo "Error: Network '$NETWORK_NAME' does not exist."
  echo "Please create it first (e.g. docker network create -d overlay --attachable $NETWORK_NAME)"
  exit 1
fi

# Auto-detect subnet prefix of the network
echo "Detecting IP subnet of '$NETWORK_NAME'..."
SUBNET=$(docker network inspect "$NETWORK_NAME" --format '{{(index .IPAM.Config 0).Subnet}}' 2>/dev/null || echo "")

if [ -z "$SUBNET" ]; then
  echo "Warning: Could not automatically detect subnet for '$NETWORK_NAME'. Using default 10.0.3.x"
  IP_PREFIX="10.0.3"
else
  IP_RAW=$(echo "$SUBNET" | cut -d'/' -f1)
  OCTET1=$(echo "$IP_RAW" | cut -d'.' -f1)
  OCTET2=$(echo "$IP_RAW" | cut -d'.' -f2)
  OCTET3=$(echo "$IP_RAW" | cut -d'.' -f3)
  IP_PREFIX="${OCTET1}.${OCTET2}.${OCTET3}"
  echo "Detected subnet prefix: ${IP_PREFIX}.x"
fi

START_IP=101

echo "Starting 10 mock agents..."
for i in {1..10}
do
  SERVER_ID=$(printf "server_%05d" $i)
  IP_ADDR="${IP_PREFIX}.$((START_IP + i - 1))"
  CONTAINER_NAME="mock-agent-$i"
  
  # Remove existing container with the same name if any
  if docker ps -a --format '{{.Names}}' | grep -Eq "^${CONTAINER_NAME}$"; then
    echo "Removing existing container $CONTAINER_NAME..."
    docker rm -f "$CONTAINER_NAME" >/dev/null
  fi

  echo "Launching $CONTAINER_NAME with IP $IP_ADDR for Server ID $SERVER_ID..."
  docker run -d \
    --name "$CONTAINER_NAME" \
    --network "$NETWORK_NAME" \
    --ip "$IP_ADDR" \
    -e APP_SERVER_ID="$SERVER_ID" \
    -e APP_HEARTBEAT_URL="$HEARTBEAT_URL" \
    -e APP_HEARTBEAT_KEY="$API_KEY" \
    -e APP_CYCLE_HEARTBEAT=5000 \
    "$IMAGE_NAME"
done

echo "Successfully started 10 mock agents!"
echo "Now you can register these servers in the database with IDs server_00001 to server_00010 and their respective IPs (starting from ${IP_PREFIX}.101)."
