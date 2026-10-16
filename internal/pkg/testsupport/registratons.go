package testsupport

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/generated/postgres/public/table"
)

// InsertRegistration - вставляет регистрацию и возвращает её id
func InsertRegistration(ctx context.Context, t *testing.T, db *pgxpool.Pool, in model.Registrations) int64 {
	t.Helper()
	stmt := table.Registrations.INSERT(
		table.Registrations.AllColumns.Except(table.Registrations.ID),
	).MODEL(
		in,
	).RETURNING(
		table.Registrations.ID,
	)

	query, args := stmt.Sql()
	var id int64
	err := db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		t.Fatalf("can't insert into registrations: %v", err)
	}

	return id
}
