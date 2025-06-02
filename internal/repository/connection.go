package repository

import (

"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewDB(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.DSN())
	if err != nil {
			return nil, err
	}
	if err := db.Ping(); err != nil {
			return nil, err
	}
	return db, nil
}

