package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/sayhellolexa/order-service/internal/kafka"
	model "github.com/sayhellolexa/order-service/internal/model"
)

func randomOrder() model.Order {
	now := time.Now().UTC()
	orderUID := uuid.New().String()

	return model.Order{
		OrderUID:    orderUID,
		TrackNumber: fmt.Sprintf("TRACK-%d", rand.Intn(100000)),
		Entry:       "WBIL",
		Delivery: model.Delivery{
			Name:    randomName(),
			Phone:   fmt.Sprintf("+97200%d", rand.Intn(9999999)),
			Zip:     fmt.Sprintf("%06d", rand.Intn(999999)),
			City:    "Kiryat Mozkin",
			Address: fmt.Sprintf("Street %d, House %d", rand.Intn(50), rand.Intn(100)),
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: model.Payment{
			Transaction:  orderUID,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       float64(rand.Intn(5000) + 500),
			PaymentDt:    now.Unix(),
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   float64(rand.Intn(2000) + 100),
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      rand.Intn(9999999),
				TrackNumber: fmt.Sprintf("TRACK-%d", rand.Intn(100000)),
				Price:       float64(rand.Intn(1000) + 100),
				Rid:         uuid.New().String(),
				Name:        randomProduct(),
				Sale:        rand.Intn(50),
				Size:        fmt.Sprintf("%d", rand.Intn(5)),
				TotalPrice:  float64(rand.Intn(2000) + 100),
				NmID:        rand.Intn(9999999),
				Brand:       randomBrand(),
				Status:      202,
			},
		},
		Locale:            "ru",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		ShardKey:          fmt.Sprintf("%d", rand.Intn(10)),
		SmID:              rand.Intn(100),
		DateCreated:       now,
		OofShard:          fmt.Sprintf("%d", rand.Intn(5)),
	}
}

func randomName() string {
	names := []string{"Ivan Ivanov", "Petr Petrov", "Sergey Sidorov", "Anna Smirnova", "Test Testov"}
	return names[rand.Intn(len(names))]
}

func randomProduct() string {
	products := []string{"Sneakers", "T-Shirt", "Jeans", "Hat", "Backpack", "Mascaras"}
	return products[rand.Intn(len(products))]
}

func randomBrand() string {
	brands := []string{"Nike", "Adidas", "Puma", "Reebok", "NewBalance"}
	return brands[rand.Intn(len(brands))]
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable not set")
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		log.Fatal("KAFKA_TOPIC environment variable not set")
	}

	producer, err := kafka.NewProducer([]string{brokers})
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	// Генерация N сообщений
	for range 10 {
		order := randomOrder()

		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("failed to marshal order: %v", err)
			continue
		}

		if err := producer.Produce(topic, data); err != nil {
			log.Printf("failed to produce message: %v", err)
		} else {
			log.Printf("Produced order %s", order.OrderUID)
		}

		time.Sleep(1 * time.Second) // задержка, чтобы не спамить слишком быстро
	}
}
