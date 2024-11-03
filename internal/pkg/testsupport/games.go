package testsupport

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TruncateGames(t *testing.T, db *pgxpool.Pool) {
	query := "TRUNCATE games"
	_, err := db.Exec(context.Background(), query)
	if err != nil {
		t.Errorf("can't truncate registrations: %v", err.Error())
	}
}
