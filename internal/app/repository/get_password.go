package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (r repository) GetPassword(ctx context.Context, tx pgx.Tx, name string) (string, error) {
	query := `
    SELECT password FROM users WHERE name = $1;
    `
	var res string
	err := tx.QueryRow(ctx, query, name).Scan(&res)
	if err != nil {
		return "", fmt.Errorf("can't select password: %w", err)
	}

	return res, nil
}
