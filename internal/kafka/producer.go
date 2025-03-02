package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"payment-gateway/configs/logger"
)

type Producer interface {
	PublishMessage(ctx context.Context, topic string, message []byte) error
	Close() error
}

var _ Producer = (*KafkaProducer)(nil)

type KafkaProducer struct {
	writer *kafka.Writer
}

func NewProducer(brokerURL string) Producer {
	p := &KafkaProducer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokerURL),
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
			BatchTimeout:           10 * time.Millisecond,
		},
	}

	return p
}

func (kp *KafkaProducer) PublishMessage(ctx context.Context, topic string, message []byte) error {
	if kp.writer == nil {
		logger.Error("Kafka writer is nil, cannot publish to Kafka")
		return fmt.Errorf("kafka writer is not initialized")
	}

	kafkaMessage := kafka.Message{
		Value: message,
		Topic: topic,
	}

	err := kp.writer.WriteMessages(ctx, kafkaMessage)
	if err != nil {
		logger.Error("Error publishing to Kafka", "error", err, "topic", topic)
		return err
	}

	logger.Info("Message successfully published to Kafka", "topic", topic)
	return nil
}

func (kp *KafkaProducer) Close() error {
	if kp.writer != nil {
		return kp.writer.Close()
	}
	return nil
}
