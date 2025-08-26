package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func (s *server) configureRoutes() {
	s.router.Use(jsonHeaderMiddleware) 

	s.router.HandleFunc("/orders/{order_uid}", s.getOrderHandler).Methods(http.MethodGet)
}

func (s *server) getOrderHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	vars := mux.Vars(r)
	id := vars["order_uid"]

	order, err := s.cache.Get(ctx, id)
	if err != nil && err != redis.Nil {
		log.Printf("Error getting order from cache: %v", err)
	}
	
	if order != nil {
		log.Printf("Order found in cache: %v", order)
		if err := json.NewEncoder(w).Encode(order); err != nil {
			log.Printf("Error encoding JSON: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return 
	}

	log.Printf("Order %s not found in cache, checking database", id)

	order, err = s.pgRepo.GetOrderById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("Order %s not found in database", id)
			http.Error(w, "Order not found in db", http.StatusNotFound)
			return
		}
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()

		if err := s.cache.Set(ctx, order, time.Hour * 72); err != nil {
			log.Printf("Error caching order: %v", err)
		} else {
			log.Printf("Order %s successfully cached", order.OrderUID)
		}
	}()

	log.Printf("Order %s found in database", id)

	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
