package testsupport

import (
	"context"
	"testing"

	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/generated/postgres/public/table"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertRegistration(t *testing.T, db *pgxpool.Pool, in model.Registrations) {
	t.Helper()
	stmt := table.Registrations.INSERT(
		table.Registrations.AllColumns,
	).MODEL(
		in,
	)

	query, args := stmt.Sql()
	_, err := db.Exec(context.Background(), query, args...)
	if err != nil {
		t.Errorf("can't insert into registrations: %v", err.Error())
	}
}

func TruncateRegistrations(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	query := "TRUNCATE registrations"
	_, err := db.Exec(context.Background(), query)
	if err != nil {
		t.Errorf("can't truncate registrations: %v", err.Error())
	}
}
