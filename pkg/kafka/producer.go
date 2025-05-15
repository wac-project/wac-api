package kafka

import (
  "context"
  "log"

  "github.com/segmentio/kafka-go"
)

var Writer *kafka.Writer

// Init wires up the global Writer
func Init(brokers []string, topic string) {
  Writer = kafka.NewWriter(kafka.WriterConfig{
    Brokers: brokers,
    Topic:   topic,
  })
}

// Send publishes one message asynchronously (fire-and-forget)
func Send(ctx context.Context, key, value []byte) {
  go func() {
    if err := Writer.WriteMessages(ctx,
      kafka.Message{Key: key, Value: value},
    ); err != nil {
      log.Printf("⚠️ kafka send error: %v", err)
    }
  }()
}
