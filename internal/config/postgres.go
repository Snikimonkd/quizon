package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"quizon/internal/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

const retries int = 3
const retryTimeout time.Duration = time.Second * 5

const dsnTemplate string = `postgres://postgres:%s@postgres:5432/postgres?sslmode=disable`

// ConnectToPostgres - подключается к postgres
func ConnectToPostgres(ctx context.Context) (*pgxpool.Pool, error) {
	pass := os.Getenv("POSTGRES_PASSWORD")
	if pass == "" {
		return nil, fmt.Errorf("can't get db password from env variable")
	}

	dsn := fmt.Sprintf(dsnTemplate, pass)
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
