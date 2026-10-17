package testsupport

import (
	"context"
	"testing"

	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/generated/postgres/public/table"

	"github.com/go-jet/jet/v2/postgres"
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

	// логины в тестах фейковые и не гарантированно уникальные: убираем за собой,
	// чтобы следующий прогон не упёрся в уникальный индекс по login
	t.Cleanup(func() {
		del := table.Admins.DELETE().WHERE(table.Admins.Login.EQ(postgres.String(m.Login)))
		query, args := del.Sql()
		_, err := db.Exec(context.WithoutCancel(ctx), query, args...)
		if err != nil {
			t.Errorf("can't delete admin: %v", err)
		}
	})
}
