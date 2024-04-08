package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
	"github.com/jackc/pgx/v5"
)

func (r repository) CheckCookie(ctx context.Context, value string) (time.Time, error) {
	query := `
    SELECT expires FROM cookies WHERE value = $1;
    `
	var res time.Time
	err := r.db.QueryRow(ctx, query, value).Scan(&res)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, model.ErrNotFound
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("can't check cookie: %w", err)
	}

	return res, nil
}
