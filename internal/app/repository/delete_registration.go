package repository

import (
	"context"
	"fmt"
)

func (r repository) DeleteRegistration(ctx context.Context, id int64) error {
	query := `
    UPDATE registrations SET deleted = true WHERE id = $1;
    `
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("can't delete registration: %w", err)
	}

	return nil
}
