package repository

import (
	"context"
	"fmt"
)

// GetCounter - получить каунтер
func (r repository) GetCounter(ctx context.Context) (int64, error) {
	query := `SELECT last_value FROM index_counter;`

	var res int64
	err := r.db.QueryRow(ctx, query).Scan(&res)
	if err != nil {
		return 0, fmt.Errorf("can't select next counter value: %w", err)
	}

	return res, nil
}
