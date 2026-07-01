package mq

import (
	"log/slog"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func InitKafkaTopics(brokerURL string) error {
	conn, err := kafka.Dial("tcp", brokerURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topics := []string{"ping", "mail", "ping_res", "heartbeat", "server_events"}
	var topicConfigs []kafka.TopicConfig
	for _, topic := range topics {
		topicConfigs = append(topicConfigs, kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     5,
			ReplicationFactor: 1,
			ConfigEntries: []kafka.ConfigEntry{
				{
					ConfigName:  "retention.ms",
					ConfigValue: "3600000",
				},
			},
		})
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		slog.Warn("kafka create topics returned error (may already exist)", "err", err)
	} else {
		slog.Info("Kafka topics created successfully")
	}

	return nil
}
