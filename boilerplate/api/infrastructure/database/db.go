package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/yourorg/boilerplate/infrastructure/config"
)

// NewDB membuat koneksi database PostgreSQL menggunakan sqlx.
func NewDB(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return db, nil
}
