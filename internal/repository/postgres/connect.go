package postgres

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/pressly/goose/v3"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(settings Settings) (*sql.DB, error) {
	sqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		settings.Host, settings.Port, settings.User, settings.Pass, settings.Name, settings.SslMode,
	)

	db, err := sql.Open("pgx", sqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Printf("Database connection was created: %s \n", sqlInfo)

	// Дропнуть структуру
	if settings.Reload {
		log.Print("Start reloading database \n")
		err = goose.DownTo(db, "./migrations", 0)
		if err != nil {
			return db, fmt.Errorf("error down to: %w", err)
		}
	}

	// Создать заново
	log.Print("Start migrating database \n")
	err = goose.Up(db, "./migrations")
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db, nil
}