package testsupport

import (
	"context"
	"sync"
	"testing"

	"quizon/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	db *pgxpool.Pool
	s  sync.Once
)

const dsn string = `postgres://postgres:some_password@localhost:5432/postgres?sslmode=disable`

// ConnectToTestPostgres - подключиться к базе в тестах
func ConnectToTestPostgres(ctx context.Context, t *testing.T) *pgxpool.Pool {
	t.Helper()
	s.Do(func() {
		var err error
		db, err = config.ConnectToPostgresByDSN(ctx, dsn)
		if err != nil {
			t.Fatalf("can't connect to db: %v", err)
		}
	})

	return db
}
