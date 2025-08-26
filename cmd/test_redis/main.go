package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	cacheRepo "github.com/sayhellolexa/order-service/internal/repository/cache"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		err = fmt.Errorf("error loading .env file: %w", err)
		log.Fatal(err)
	}

	redisAddress := os.Getenv("REDIS_URL")
	if redisAddress == "" {
		log.Fatal("REDIS_URL environment variable not set")
	}

	// Создаем клиент Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "",
		DB:       0,
	})

	// Проверяем подключение к Redis
	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Printf("Connected to Redis: %s\n", pong)

	// Создаем репозиторий кэша
	cache := cacheRepo.NewRedisCacheRepository(rdb, nil)

	// Пытаемся получить данные из кэша
	orderUID := "b563feb7b2b84b556test"
	fmt.Printf("Attempting to get order %s from cache...\n", orderUID)
	
	order, err := cache.Get(ctx, orderUID)
	if err != nil {
		log.Printf("Error getting order from cache: %v", err)
	} else if order == nil {
		fmt.Printf("Order %s not found in cache\n", orderUID)
	} else {
		fmt.Printf("Successfully retrieved order %s from cache\n", order.OrderUID)
		// fmt.Printf("Order details: %+v\n", order)
	}
	
	// Проверяем все ключи в Redis
	fmt.Println("Checking all keys in Redis...")
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		log.Printf("Error getting keys from Redis: %v", err)
	} else {
		fmt.Printf("Found %d keys in Redis:\n", len(keys))
		for _, key := range keys {
			fmt.Printf("  - %s\n", key)
			
			// Получаем значение ключа
			val, err := rdb.Get(ctx, key).Result()
			if err != nil {
				log.Printf("Error getting value for key %s: %v", key, err)
			} else {
				fmt.Printf("    Value length: %d bytes\n", len(val))
			}
		}
	}
	
	// Ждем немного, чтобы увидеть TTL ключей
	fmt.Println("Checking TTL for keys...")
	for _, key := range keys {
		ttl, err := rdb.TTL(ctx, key).Result()
		if err != nil {
			log.Printf("Error getting TTL for key %s: %v", key, err)
		} else {
			fmt.Printf("  - %s: TTL = %v\n", key, ttl)
		}
	}
}
