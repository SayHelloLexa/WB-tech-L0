package kafka

import (
	"errors"
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const(
	flushTimeout = 5000 // ms
)

var errUnknownType = errors.New("unknown event type")

type Producer struct {
	producer *kafka.Producer
}

func NewProducer (address []string) (*Producer, error) {
	conf := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(address, ","), // Список брокеров, в которые будем отправлять сообщения
	}

	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, fmt.Errorf("error with new producer: %w", err)
	}
	
	return &Producer{producer: p}, nil
}

func (p *Producer) Produce(topic string, data []byte) error {
	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: data, // <-- сразу передаём байты
	}

	kafkaChan := make(chan kafka.Event, 1)

	if err := p.producer.Produce(kafkaMsg, kafkaChan); err != nil {
		return fmt.Errorf("error with produce: %w", err)
	}

	e := <-kafkaChan
	switch ev := e.(type) {
	case *kafka.Message:
		return nil
	case kafka.Error:
		return ev
	default:
		return errUnknownType
	}
}

// Flush используется, чтобы заблокировать close пока
// не доотправятся все сообщения или не сработает таймаут
func (p *Producer) Close() {
	p.producer.Flush(flushTimeout)
	p.producer.Close()
}