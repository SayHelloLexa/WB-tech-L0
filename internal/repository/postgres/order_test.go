package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	model "github.com/sayhellolexa/order-service/internal/model"
)

var (
	db *sql.DB
	mock sqlmock.Sqlmock
	ctx context.Context
)

func TestMain(m *testing.M) {
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		log.Fatalf("error creating sqlmock: %s", err)
	}
	defer db.Close()

	ctx = context.Background()

	code := m.Run()

	os.Exit(code)
}

func createTestRepository() *OrderRepository {
	return NewOrderRepository(db)
}

func TestOrderRepository_GetOrderById(t *testing.T) {
	repo := createTestRepository()

	data, err := os.ReadFile("testdata.json")
	if err != nil {
		log.Fatal(err)
	}
	
	var testOrder model.Order
	err = json.Unmarshal(data, &testOrder)
	if err != nil {
		log.Fatal(err)
	}

	testTable := []struct {
		name       string
		orderUID   string
		mockDB     func(id string)
		expected   *model.Order
		wantErr    bool
	}{
		{
			name:     "Order found in DB",
			orderUID: "test1",
			mockDB: func(id string) {
				// Мокируем основной запрос
				mainRows := sqlmock.NewRows([]string{
					"order_uid", "track_number", "entry", "locale", "internal_signature",
					"customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard",
					"name", "phone", "zip", "city", "address", "region", "email",
					"transaction", "request_id", "currency", "provider", "amount",
					"payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee",
				}).
					AddRow(
						testOrder.OrderUID, testOrder.TrackNumber, testOrder.Entry, testOrder.Locale,
						testOrder.InternalSignature, testOrder.CustomerID, testOrder.DeliveryService,
						testOrder.ShardKey, testOrder.SmID, testOrder.DateCreated, testOrder.OofShard,
						testOrder.Delivery.Name, testOrder.Delivery.Phone, testOrder.Delivery.Zip,
						testOrder.Delivery.City, testOrder.Delivery.Address, testOrder.Delivery.Region,
						testOrder.Delivery.Email, testOrder.Payment.Transaction, testOrder.Payment.RequestID,
						testOrder.Payment.Currency, testOrder.Payment.Provider,
						testOrder.Payment.Amount, testOrder.Payment.PaymentDt, testOrder.Payment.Bank,
						testOrder.Payment.DeliveryCost, testOrder.Payment.GoodsTotal, testOrder.Payment.CustomFee,
					)

				mock.ExpectQuery(
					`SELECT (.+) FROM orders o JOIN deliveries d ON .+ JOIN payments p ON .+ WHERE o\.order_uid = \$1`,
				).WithArgs(id).WillReturnRows(mainRows)

				itemRows := sqlmock.NewRows([]string{
					"chrt_id", "track_number", "price", "rid", "name",
					"sale", "size", "total_price", "nm_id", "brand", "status",
				})
				for _, item := range testOrder.Items {
					itemRows.AddRow(
						item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
						item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status,
					)
				}

				mock.ExpectQuery(
					`SELECT (.+) FROM items WHERE order_uid = \$1`,
				).WithArgs(id).WillReturnRows(itemRows)
			},
			expected: &testOrder,
			wantErr:  false,
		},
		{
			name:     "Order not found",
			mockDB: func(id string) {
				mock.ExpectQuery(
					`SELECT (.+) FROM orders o JOIN deliveries d ON .+ JOIN payments p ON .+ WHERE o\.order_uid = \$1`,
				).WithArgs(id).WillReturnError(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "DB error",
			orderUID: "test4",
			mockDB: func(id string) {
				mock.ExpectQuery(
					`SELECT (.+) FROM orders o JOIN deliveries d ON .+ JOIN payments p ON .+ WHERE o\.order_uid = \$1`,
				).WithArgs(id).WillReturnError(fmt.Errorf("database error"))
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockDB(testCase.orderUID)

			got, err := repo.GetOrderById(ctx, testCase.orderUID)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, got, testCase.expected)
			}
		})
	}
}