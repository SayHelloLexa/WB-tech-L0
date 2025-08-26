-- +goose Up
CREATE TABLE IF NOT EXISTS payments (
   order_uid VARCHAR(50) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
   transaction VARCHAR(100) NOT NULL,
   request_id VARCHAR(50),
   currency VARCHAR(10) NOT NULL,
   provider VARCHAR(50) NOT NULL,
   amount DECIMAL(12, 2) NOT NULL,
   payment_dt BIGINT NOT NULL,
   bank VARCHAR(50) NOT NULL,
   delivery_cost DECIMAL(12, 2) NOT NULL,
   goods_total DECIMAL(12, 2) NOT NULL,
   custom_fee DECIMAL(12, 2) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS payments