package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	cache "github.com/sayhellolexa/order-service/internal/domain/cache"
	order "github.com/sayhellolexa/order-service/internal/domain/order"
	model "github.com/sayhellolexa/order-service/internal/model"
)

type Handler struct {
	orderRepository order.Repository
	cacheRepository cache.Repository
}

func NewHandler(orderRepository order.Repository, cacheRepository cache.Repository) *Handler {
	return &Handler{orderRepository: orderRepository, cacheRepository: cacheRepository}
}

func (h *Handler) HandleMessage(ctx context.Context, message []byte, offset kafka.Offset) error {
	log.Printf("Received message from Kafka with offset: %d", offset)
	
	err := h.orderRepository.SaveOrder(ctx, message)
	if err != nil {
		log.Printf("Error saving order to database: %v", err)
		return fmt.Errorf("error with SaveOrder on kafka handler: %w", err)
	}

	var order model.Order 
	err = json.Unmarshal(message, &order)
	if err != nil {
		log.Printf("Error unmarshaling order from Kafka message: %v", err)
		return fmt.Errorf("error with Unmarshal on kafka handler: %w", err)
	}

	log.Printf("Successfully unmarshaled order with ID: %s", order.OrderUID)
	
	err = h.cacheRepository.Set(ctx, &order, time.Hour * 72)
	if err != nil {
		log.Printf("Error setting order in cache: %v", err)
		return fmt.Errorf("error with set cache on kafka handler: %w", err)
	}

	log.Printf("Message from Kafka with offset save to db and cache: %d, order ID: %s", offset, order.OrderUID)

	return nil
}
