#!/bin/bash

echo "Stopping and removing 10 mock agents..."
for i in {1..10}
do
  CONTAINER_NAME="mock-agent-$i"
  if docker ps -a --format '{{.Names}}' | grep -Eq "^${CONTAINER_NAME}$"; then
    echo "Stopping and removing $CONTAINER_NAME..."
    docker rm -f "$CONTAINER_NAME" >/dev/null
  else
    echo "$CONTAINER_NAME is not running."
  fi
done

echo "Clean up completed!"
