-- +goose Up
CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(50) NOT NULL REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id BIGINT NOT NULL,
    track_number VARCHAR(50) NOT NULL,
    price DECIMAL(12, 2) NOT NULL,
    rid VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL, 
    sale INTEGER NOT NULL, 
    size VARCHAR(10) NOT NULL,
    total_price DECIMAL(12, 2) NOT NULL,
    nm_id BIGINT NOT NULL,
    brand VARCHAR(100) NOT NULL,
    status INTEGER NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS items