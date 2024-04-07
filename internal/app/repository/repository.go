package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

// New - конструктор
func New(db *pgxpool.Pool) repository {
	return repository{
		db: db,
	}
}
