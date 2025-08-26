package domain

import (
	"context"

	model "github.com/sayhellolexa/order-service/internal/model"
)

// Абстракция, которая определяет методы для работы с заказами
type Repository interface {
	GetOrderById(ctx context.Context, id string) (*model.Order, error)
	SaveOrder(ctx context.Context, message []byte) error
}