package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/sayhellolexa/order-service/internal/kafka"
	"github.com/sayhellolexa/order-service/internal/kafka/handler"
	"github.com/sayhellolexa/order-service/internal/repository/cache"
	"github.com/sayhellolexa/order-service/internal/repository/postgres"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		err = fmt.Errorf("error loading .env file: %w", err)
		log.Fatal(err)
	}

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable not set")
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		log.Fatal("KAFKA_TOPIC environment variable not set")
	}

	consumerGroup := os.Getenv("KAFKA_CONSUMER_GROUP")
	if consumerGroup == "" {
		log.Fatal("KAFKA_CONSUMER_GROUP environment variable not set")
	}

	dataBaseUrl := os.Getenv("DATABASE_URL")
	if dataBaseUrl == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	redisAddress := os.Getenv("REDIS_URL")
	if redisAddress == "" {
		log.Fatal("REDIS_URL environment variable not set")
	}

	db, err := sql.Open("pgx", dataBaseUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddress,
		Password: "",
		DB: 0,
	})

	err = rdb.Ping(context.Background()).Err()
	if err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}

	cache := cache.NewRedisCacheRepository(rdb, db)	

	repo := postgres.NewOrderRepository(db)

	handler := handler.NewHandler(repo, cache)

	c, err := kafka.NewConsumer([]string{brokers}, consumerGroup, topic, handler)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	if c == nil {
		log.Fatal("Consumer is nil")
	}

	go c.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Fatal(c.Stop())
}
