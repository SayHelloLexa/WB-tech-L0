package cache

import (
	"context"
	"time"

	model "github.com/sayhellolexa/order-service/internal/model"
)

type Repository interface {
	Get(ctx context.Context, orderUID string) (*model.Order, error)
	Set(ctx context.Context, order *model.Order, ttl time.Duration) error
	Count(ctx context.Context) (int64, error) 
	GetAllOrdersIDs(ctx context.Context) ([]string, error)
	PreloadFromDatabase(ctx context.Context, batchSize int) error
}