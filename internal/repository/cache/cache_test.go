package cache_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	redismock "github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/require"

	model "github.com/sayhellolexa/order-service/internal/model"
	"github.com/sayhellolexa/order-service/internal/repository/cache"
)

func TestRedisCache_Get(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	db, _, _ := sqlmock.New()
	c := cache.NewRedisCacheRepository(rdb, db)

	ctx := context.Background()
	order := &model.Order{OrderUID: "123"}
	data, _ := json.Marshal(order)

	tests := []struct {
		name    string
		mockFn  func()
		wantErr bool
		wantNil bool
	}{
		{
			name: "cache hit",
			mockFn: func() {
				mock.ExpectGet("order:123").SetVal(string(data))
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "cache miss",
			mockFn: func() {
				mock.ExpectGet("order:123").RedisNil()
			},
			wantErr: false,
			wantNil: true,
		},
		{
			name: "redis error",
			mockFn: func() {
				mock.ExpectGet("order:123").SetErr(errors.New("boom"))
			},
			wantErr: true,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			tt.mockFn()
			got, err := c.Get(ctx, "123")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			if tt.wantNil {
				require.Nil(t, got)
			} else {
				require.NotNil(t, got)
				require.Equal(t, order.OrderUID, got.OrderUID)
			}
		})
	}
}

func TestRedisCache_Set(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	db, _, _ := sqlmock.New()
	c := cache.NewRedisCacheRepository(rdb, db)


	ctx := context.Background()
	order := &model.Order{OrderUID: "123"}
	data, _ := json.Marshal(order)


	tests := []struct {
		name string
		mockFn func()
		wantErr bool
	}{
		{
			name: "success",
			mockFn: func() {
			mock.ExpectSet("order:123", data, time.Hour).SetVal("OK")
		},
			wantErr: false,
		},
		{
			name: "redis error",
			mockFn: func() {
			mock.ExpectSet("order:123", data, time.Hour).SetErr(errors.New("boom"))
		},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			tt.mockFn()
			err := c.Set(ctx, order, time.Hour)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}