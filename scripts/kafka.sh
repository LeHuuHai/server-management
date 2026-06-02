#!/bin/bash

set -e

BOOTSTRAP=sm_kafka:29092

echo "Waiting for Kafka..."

until /opt/kafka/bin/kafka-topics.sh --bootstrap-server $BOOTSTRAP --list; do
  sleep 2
done

echo "Kafka is ready. Creating topics..."

echo "Creating topic: ping"
/opt/kafka/bin/kafka-topics.sh --create --if-not-exists \
  --topic ping \
  --bootstrap-server $BOOTSTRAP \
  --partitions 5 \
  --replication-factor 1 \
  --config retention.ms=3600000

echo "Creating topic: mail"
/opt/kafka/bin/kafka-topics.sh --create --if-not-exists \
  --topic mail \
  --bootstrap-server $BOOTSTRAP \
  --partitions 5 \
  --replication-factor 1 \
  --config retention.ms=3600000

echo "Creating topic: ping_res"
/opt/kafka/bin/kafka-topics.sh --create --if-not-exists \
  --topic ping_res \
  --bootstrap-server $BOOTSTRAP \
  --partitions 5 \
  --replication-factor 1 \
  --config retention.ms=3600000

echo "Creating topic: heartbeat"
/opt/kafka/bin/kafka-topics.sh --create --if-not-exists \
  --topic heartbeat \
  --bootstrap-server $BOOTSTRAP \
  --partitions 5 \
  --replication-factor 1 \
  --config retention.ms=3600000

echo "All topics created successfully."