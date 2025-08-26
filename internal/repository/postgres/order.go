package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	model "github.com/sayhellolexa/order-service/internal/model"
)

type OrderRepository struct {
	db *sql.DB
}

// Конструктор для нового экземпляра OrderRepository
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetOrderById(ctx context.Context, id string) (*model.Order, error) {
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

	err := r.db.QueryRowContext(ctx, orderQuery, id).Scan(
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

	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
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

// TODO: Добавить валидацию данных перед сохранением в БД
func (r *OrderRepository) SaveOrder(ctx context.Context, message []byte) error {
	var orderMsg model.Order
	if err := json.Unmarshal(message, &orderMsg); err != nil {
		return fmt.Errorf("error unmarshaling message: %w", err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("failed to rollback transaction: %v", rbErr)
			}
		}
	}()

	orderInsertQuery := `INSERT INTO orders (
		order_uid, track_number, entry, locale, internal_signature, customer_id, 
		delivery_service, shardkey, sm_id, date_created, oof_shard
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = tx.Exec(orderInsertQuery, 
		orderMsg.OrderUID, orderMsg.TrackNumber, orderMsg.Entry, orderMsg.Locale,
		orderMsg.InternalSignature, orderMsg.CustomerID, orderMsg.DeliveryService,
		orderMsg.ShardKey, orderMsg.SmID, orderMsg.DateCreated, orderMsg.OofShard,
	)
	if err != nil {
		return fmt.Errorf("error with order insert query: %w", err)
	}

	deliveryInsertQuery := `INSERT INTO deliveries (
		order_uid, name, phone, zip, city, address, region, email
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err = tx.ExecContext(ctx, deliveryInsertQuery,
		orderMsg.OrderUID, orderMsg.Delivery.Name, orderMsg.Delivery.Phone,
		orderMsg.Delivery.Zip, orderMsg.Delivery.City, orderMsg.Delivery.Address,
		orderMsg.Delivery.Region, orderMsg.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("error with delivery insert query: %w", err)
	}

	paymentInsertQuery := `INSERT INTO payments (
		order_uid, transaction, request_id, currency, provider, amount,
		payment_dt, bank, delivery_cost, goods_total, custom_fee
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = tx.ExecContext(ctx, paymentInsertQuery,
		orderMsg.OrderUID, orderMsg.Payment.Transaction, orderMsg.Payment.RequestID,
		orderMsg.Payment.Currency, orderMsg.Payment.Provider, orderMsg.Payment.Amount,
		orderMsg.Payment.PaymentDt, orderMsg.Payment.Bank, orderMsg.Payment.DeliveryCost,
		orderMsg.Payment.GoodsTotal, orderMsg.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("error with payment insert query: %w", err)
	}

	itemInserQuery := `INSERT INTO items (
		order_uid, chrt_id, track_number, price, rid, name, sale, size,
		total_price, nm_id, brand, status
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	for _, item := range orderMsg.Items {
		_, err = tx.ExecContext(ctx, itemInserQuery,
			orderMsg.OrderUID, item.ChrtID, item.TrackNumber, item.Price,
			item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status,
		)
	}
	if err != nil {
		return fmt.Errorf("error with item insert query: %w", err)
	}
	
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing trans: %w", err)
	}

	return nil
}
