package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	sessionTimeout = 10000
	noTimeout      = -1
)

type Handler interface {
	HandleMessage(ctx context.Context, message []byte, offset kafka.Offset) error
}

type Consumer struct {
	consumer *kafka.Consumer
	handler  Handler
	stop     bool
}

func NewConsumer(address []string, consumerGroup string, topic string, handler Handler) (*Consumer, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(address, ","),
		"group.id": consumerGroup,
		"session.timeout.ms": sessionTimeout,
		"enable.auto.offset.store": false,
		"enable.auto.commit": true,
		"auto.commit.interval.ms": 5000,
		"auto.offset.reset": "earliest",
	}
	c, err := kafka.NewConsumer(conf)
	if err != nil {
		return nil, fmt.Errorf("error with new consumer: %w", err)
	}

	if err := c.Subscribe(topic, nil); err != nil {
		return nil, fmt.Errorf("error with subscribe: %w", err)
	}

	return &Consumer{consumer: c, handler: handler}, nil
}

func (c *Consumer) Start() {
	ctx := context.Background()

	for {
		if c.stop {
			break
		}

		kafkaMsg, err := c.consumer.ReadMessage(noTimeout)
		if err != nil {
			fmt.Printf("error with read message: %v\n", err)
			continue
		}
		if kafkaMsg == nil {
			continue
		}

		if err := c.handler.HandleMessage(ctx, kafkaMsg.Value, kafkaMsg.TopicPartition.Offset); err != nil {
			fmt.Printf("error with handle message: %v\n", err)
			continue
		}

		if _, err := c.consumer.StoreMessage(kafkaMsg); err != nil {
			fmt.Printf("error with store message: %v\n", err)
			continue
		}
	}
}

func (c *Consumer) Stop() error {
	c.stop = true
	return c.consumer.Close()
}
