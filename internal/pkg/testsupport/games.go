package testsupport

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/generated/postgres/public/table"
)

// InsertGame - вставляет игру и возвращает её id
func InsertGame(ctx context.Context, t *testing.T, db *pgxpool.Pool, in model.Games) int64 {
	t.Helper()
	stmt := table.Games.INSERT(
		table.Games.AllColumns.Except(table.Games.ID),
	).MODEL(
		in,
	).RETURNING(
		table.Games.ID,
	)

	query, args := stmt.Sql()
	var id int64
	err := db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		t.Fatalf("can't insert into games: %v", err)
	}

	return id
}
