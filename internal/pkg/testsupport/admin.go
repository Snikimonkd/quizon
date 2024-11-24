package testsupport

import (
	"context"
	"testing"

	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/generated/postgres/public/table"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertIntoAdmins(ctx context.Context, t *testing.T, db *pgxpool.Pool, m model.Admins) {
	t.Helper()
	stmt := table.Admins.INSERT(
		table.Admins.AllColumns,
	).MODEL(
		m,
	)

	query, args := stmt.Sql()
	_, err := db.Exec(ctx, query, args...)
	if err != nil {
		t.Errorf("can't insert into admins: %v", err)
	}
}
