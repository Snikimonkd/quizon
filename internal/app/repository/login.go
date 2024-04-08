package repository

import (
	"context"
	"fmt"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
	"github.com/jackc/pgx/v5"
)

func (r repository) Login(ctx context.Context, tx pgx.Tx, cookie model.Cookie) error {
	query := `
    INSERT INTO cookies(admin_name, value, expires) VALUES($1, $2, $3);
    `
	_, err := tx.Exec(ctx, query, cookie.AdminName, cookie.Value, cookie.Expires)
	if err != nil {
		return fmt.Errorf("can't insert into cookies: %w", err)
	}
	return nil
}
