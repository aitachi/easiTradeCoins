//Author:Aitachi
//Email:44158892@qq.com
//Date: 11-02-2025 17

package messaging

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	Brokers []string
	Topic   string
}

// KafkaProducer wraps kafka producer
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer creates a new Kafka producer
// This is a backup configuration - not used in testing
func NewKafkaProducer(cfg *KafkaConfig) *KafkaProducer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.Topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaProducer{
		writer: w,
	}
}

// PublishTrade publishes a trade event to Kafka
func (p *KafkaProducer) PublishTrade(ctx context.Context, trade interface{}) error {
	data, err := json.Marshal(trade)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Value: data,
	})
}

// Close closes the Kafka producer
func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

// KafkaConsumer wraps kafka consumer
type KafkaConsumer struct {
	reader *kafka.Reader
}

// NewKafkaConsumer creates a new Kafka consumer
// This is a backup configuration - not used in testing
func NewKafkaConsumer(cfg *KafkaConfig, groupID string) *KafkaConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: groupID,
	})

	return &KafkaConsumer{
		reader: r,
	}
}

// ReadMessage reads a message from Kafka
func (c *KafkaConsumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

// Close closes the Kafka consumer
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

// Example usage (commented out - backup configuration only)
/*
func ExampleKafkaUsage() {
	// Producer
	producer := NewKafkaProducer(&KafkaConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "trades",
	})
	defer producer.Close()

	// Publish trade
	ctx := context.Background()
	err := producer.PublishTrade(ctx, map[string]interface{}{
		"symbol": "BTC_USDT",
		"price":  "45000",
		"qty":    "0.1",
	})
	if err != nil {
		log.Printf("Failed to publish: %v", err)
	}

	// Consumer
	consumer := NewKafkaConsumer(&KafkaConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "trades",
	}, "trade-processor")
	defer consumer.Close()

	// Read messages
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			break
		}
		log.Printf("Received: %s", string(msg.Value))
	}
}
*/
