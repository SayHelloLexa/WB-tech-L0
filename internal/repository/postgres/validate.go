package postgres

import (
	"fmt"

	model "github.com/sayhellolexa/order-service/internal/model"
)

func validateOrder(o *model.Order) error {
	if o.OrderUID == "" {
		return fmt.Errorf("order_uid is required")
	}
	if o.TrackNumber == "" {
		return fmt.Errorf("track_number is required")
	}
	if o.CustomerID == "" {
		return fmt.Errorf("customer_id is required")
	}
	if o.Delivery.Name == "" || o.Delivery.Address == "" {
		return fmt.Errorf("delivery name and address are required")
	}
	if o.Payment.Transaction == "" {
		return fmt.Errorf("payment transaction is required")
	}
	if o.Payment.Amount < 0 {
		return fmt.Errorf("payment amount cannot be negative")
	}

	for i, item := range o.Items {
		if item.ChrtID == 0 {
			return fmt.Errorf("item[%d]: chrt_id is required", i)
		}
		if item.Name == "" {
			return fmt.Errorf("item[%d]: name is required", i)
		}
		if item.Price < 0 || item.TotalPrice < 0 {
			return fmt.Errorf("item[%d]: price/total_price cannot be negative", i)
		}
	}
	return nil
}
