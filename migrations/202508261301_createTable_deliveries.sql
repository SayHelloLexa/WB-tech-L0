-- +goose Up
CREATE TABLE IF NOT EXISTS deliveries (
    order_uid VARCHAR(50) PRIMARY KEY REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    phone VARCHAR(20) NOT NULL,
    zip VARCHAR(20) NOT NULL,
    city VARCHAR(100) NOT NULL,
    address VARCHAR(200) NOT NULL,
    region VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS deliveries