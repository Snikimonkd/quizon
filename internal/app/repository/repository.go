package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

// NewRepository - конструктор
func NewRepository(db *pgxpool.Pool) repository {
	return repository{
		db: db,
	}
}
