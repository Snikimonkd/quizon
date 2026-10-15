package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"quizon/internal/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	retries      int           = 3
	retryTimeout time.Duration = time.Second * 5
)

// ConnectToPostgres - подключается к postgres
func ConnectToPostgres(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		return nil, fmt.Errorf("empty dsn in env variable")
	}

	return ConnectToPostgresByDSN(ctx, dsn)
}

func ConnectToPostgresByDSN(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("can't connect to db: %w", err)
	}

	for range retries {
		err := db.Ping(ctx)
		if err == nil {
			return db, nil
		}

		logger.Errorf("can't ping db: %v", err)
		time.Sleep(retryTimeout)
	}

	return nil, fmt.Errorf("failed to ping db after %v retries", retries)
}
