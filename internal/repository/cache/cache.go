package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	model "github.com/sayhellolexa/order-service/internal/model"
)

type RedisCache struct {
	client *redis.Client
	db *sql.DB
}

func NewRedisCacheRepository(client *redis.Client, db *sql.DB) *RedisCache {
	return &RedisCache{client: client, db: db}
}

// Получить кеш
func (c *RedisCache) Get(ctx context.Context, orderUID string) (*model.Order, error) {
	key := fmt.Sprintf("order:%s", orderUID)
	
	log.Printf("Attempting to get order from cache with ID: %s", orderUID)
	
	val, err := c.client.Get(ctx, key).Result()
	if err == nil {
		log.Printf("Cache HIT for order ID: %s, data length: %d bytes", orderUID, len(val))
		
		var order model.Order
		if err := json.Unmarshal([]byte(val), &order); err != nil {
			log.Printf("Failed to unmarshal order %s from cache: %v", orderUID, err)
			return nil, fmt.Errorf("failed to unmarshal order from cache: %w", err)
		}
		
		log.Printf("Successfully unmarshaled order %s from cache", orderUID)
		return &order, nil
	}

	if err != redis.Nil {
		log.Printf("Redis error on GET for order %s: %v", orderUID, err)
		return nil, fmt.Errorf("redis error on GET: %w", err)
	} else {
		log.Printf("Cache MISS for order ID: %s", orderUID)
		return nil, nil
	}
}

// Установить кеш
func (c *RedisCache) Set(ctx context.Context, order *model.Order, ttl time.Duration) error {
	key := fmt.Sprintf("order:%s", order.OrderUID)
	
	log.Printf("Attempting to cache order with ID: %s", order.OrderUID)
	
	data, err := json.Marshal(order)
	if err != nil {
		log.Printf("Failed to marshal order %s: %v", order.OrderUID, err)
		return fmt.Errorf("failed to marshal order for cache: %w", err)
	}
	
	log.Printf("Successfully marshaled order %s, data length: %d bytes", order.OrderUID, len(data))
	
	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		log.Printf("Failed to set order %s in cache: %v", order.OrderUID, err)
		return fmt.Errorf("failed to set order in cache: %w", err)
	}
	
	log.Printf("Successfully cached order with ID: %s", order.OrderUID)
	
	return nil
}

func (c *RedisCache) Count(ctx context.Context) (int64, error) {
	return c.client.DBSize(ctx).Result()
}

func (c *RedisCache) GetOrderById(ctx context.Context, id string) (*model.Order, error) {
	orderQuery := `
        SELECT o.order_uid, o.track_number, o.entry, o.locale, 
	         o.internal_signature, 
             o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
             d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
             p.transaction, p.request_id, p.currency, p.provider, p.amount, 
             p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee 
       	FROM orders o
		JOIN deliveries d ON o.order_uid = d.order_uid
		JOIN payments p ON o.order_uid = p.order_uid
		WHERE o.order_uid = $1
	`
	
	itemsQuery := `
	    SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1
	`

	var order model.Order
	var delivery model.Delivery
	var payment model.Payment

	var items []model.Item

	err := c.db.QueryRowContext(ctx, orderQuery, id).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
        &order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard,
        &delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
        &payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount,
        &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	order.Delivery = delivery
	order.Payment = payment

	rows, err := c.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.Item
		err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
            &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	order.Items = items

	return &order, nil
}

func (c *RedisCache) GetAllOrdersIDs(ctx context.Context) ([]string, error) {
	query := `SELECT order_uid FROM orders ORDER BY date_created DESC`

	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all orders IDs: %w", err)
	}
	defer rows.Close()

	var orderIDs []string
	for rows.Next() {
		var orderID string
		if err := rows.Scan(&orderID); err != nil {
			return nil, fmt.Errorf("failed to scan all orders IDs: %w", err)
		}
		orderIDs = append(orderIDs, orderID)
	}

	if err = rows.Err(); err != nil {
		return orderIDs, err
	}

	return orderIDs, nil
}

// Прелзагрузка из БД
func (c *RedisCache) PreloadFromDatabase(ctx context.Context, batchSize int) error {
	log.Printf("Starting cache preloading from db...")

	if batchSize <= 0 {
		return errors.New("invalid batch size")
	}

	orderIDs, err := c.GetAllOrdersIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get order IDs: %w", err)
	}

	successCount := 0

	for _, orderID := range orderIDs {
		order, err := c.GetOrderById(ctx, orderID)
		if err != nil {
			log.Printf("Failed to get order %s: %v", orderID, err)
			continue
		}

		if order == nil {
			log.Printf("Order %s not found in db", orderID)
			continue
		}

		if err := c.Set(ctx, order, time.Hour * 72); err != nil {
			log.Printf("Failed to set order %s in cache: %v", orderID, err)
			continue
		}

		successCount++ 
	}

	log.Printf("Cache preloading completed: %d/%d orders loaded", successCount, len(orderIDs))

	return nil
}
