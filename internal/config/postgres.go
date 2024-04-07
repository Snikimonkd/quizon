package config

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

// NewPostgres - новый пул коннектов до постгреса
func NewPostgres(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := viper.GetString("pg-dsn")
	if dsn == "" {
		return nil, fmt.Errorf("empty pg dsn")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("can't connect to postgres: %w", err)
	}

	return pool, nil
}
