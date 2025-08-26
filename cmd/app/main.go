package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/sayhellolexa/order-service/internal/repository/cache"
	"github.com/sayhellolexa/order-service/internal/repository/postgres"
	"github.com/sayhellolexa/order-service/internal/server"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		err = fmt.Errorf("error loading .env file: %w", err)
		log.Fatal(err)
	}

	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		log.Fatal("SERVER_ADDR environment variable not set")
	}

	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		log.Fatal("DATABASE_NAME environment variable not set")
	}

	databaseHost := os.Getenv("DATABASE_HOST")
	if databaseHost == "" {
		log.Fatal("DATABASE_HOST environment variable not set")
	}

	databasePort := os.Getenv("DATABASE_PORT")
	if databasePort == "" {
		log.Fatal("DATABASE_PORT environment variable not set")
	}

	databaseUser := os.Getenv("DATABASE_USER")
	if databaseUser == "" {
		log.Fatal("DATABASE_USER environment variable not set")
	}

	databasePass := os.Getenv("DATABASE_PASS")
	if databasePass == "" {
		log.Fatal("DATABASE_PASS environment variable not set")
	}

	databaseSsl := os.Getenv("DATABASE_SSL_MODE")
	if databaseSsl == "" {
		log.Fatal("DATABASE_SSL_MODE environment variable not set")
	}

	settings := postgres.Settings{
		User: databaseUser,
		Pass: databasePass,
		Host:  databaseHost,
		Port: databasePort, 
		Name: databaseName,
		SslMode: databaseSsl,
		Reload: false,
	}

	redisAddress := os.Getenv("REDIS_URL")
	if redisAddress == "" {
		log.Fatal("REDIS_URL environment variable not set")
	}

	db, err := postgres.Connect(settings)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Unable to ping database: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: "", 
		DB:       0,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}

	pgRepo := postgres.NewOrderRepository(db)
	cacheRepo := cache.NewRedisCacheRepository(rdb, db)

	s := server.NewServer(pgRepo, cacheRepo)

	if err := s.Start(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Print("Server is working...")
}