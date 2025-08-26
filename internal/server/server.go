package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sayhellolexa/order-service/internal/domain/cache"
	"github.com/sayhellolexa/order-service/internal/domain/order"
)

type server struct {
	httpServer *http.Server
	router *mux.Router
	pgRepo domain.Repository
	cache cache.Repository
}

func NewServer(pgRepo domain.Repository, cacheRepo cache.Repository) *server {
	s := &server{
		router: mux.NewRouter(),
		pgRepo: pgRepo,
		cache: cacheRepo,
	}

	s.configureRoutes()
	return s
}

func (s *server) Start(addr string) error {
	s.httpServer = &http.Server{
		Addr: addr,
		Handler: s.router,
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
		defer cancel()

		count, err := s.cache.Count(ctx) // метод, который вернёт количество ключей в Redis
		if err != nil {
			log.Printf("failed to check cache: %v", err)
			return
		}

		if count == 0 {
			log.Println("Cache is empty → start preload from DB")
			if err := s.cache.PreloadFromDatabase(ctx, 100); err != nil {
				log.Printf("failed to preload cache: %v", err)
			} else {
				log.Println("Cache preloaded successfully")
			}
		} else {
			log.Printf("Cache already has %d records, preload skipped", count)
		}
	}()

	err := s.httpServer.ListenAndServe()
	if err != nil {
		return fmt.Errorf("error with start http server: %w", err)
	}

	return nil
}
