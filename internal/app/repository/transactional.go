package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

func (r repository) Transactional(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("can't begin tx: %w", err)
	}

	err = fn(ctx, tx)
	if err != nil {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil {
			log.Error().Err(err).Msg("can't rollback transaction")
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("can't commit tx: %w", err)
	}

	return nil
}
