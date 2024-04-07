package repository

import (
	"context"
	"fmt"
)

func (r repository) Inc(ctx context.Context) (int64, error) {
	query := `SELECT nextval('index_counter');`

	var res int64
	err := r.db.QueryRow(ctx, query).Scan(&res)
	if err != nil {
		return 0, fmt.Errorf("can't select next counter value: %w", err)
	}

	return res, nil
}
