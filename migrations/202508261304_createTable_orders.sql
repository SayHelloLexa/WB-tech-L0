-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(50) PRIMARY KEY,
    track_number VARCHAR(50) NOT NULL,
    entry VARCHAR(10) NOT NULL,
    locale VARCHAR(10) NOT NULL,
    internal_signature VARCHAR(100) NOT NULL,
    customer_id VARCHAR(50) NOT NULL,
    delivery_service VARCHAR(50) NOT NULL,
    shardkey VARCHAR(10) NOT NULL,
    sm_id INTEGER NOT NULL,
    date_created TIMESTAMP NOT NULL,
    oof_shard VARCHAR(10) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS orders