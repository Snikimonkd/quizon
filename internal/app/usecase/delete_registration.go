package usecase

import (
	"context"
	"errors"
	"time"

	"quizon/internal/app/repository"

	"github.com/jackc/pgx/v5"
)

type DeleteRegistrationRepository interface {
	DeleteRegistration(ctx context.Context, tx pgx.Tx, gameID int64, registrationID int64, deletedAt time.Time) error
}

// DeleteRegistration - убирает команду из игры. Под той же блокировкой игры, что и регистрация,
// чтобы новая команда не посчитала места, пока освобождается чужое.
func (u usecase) DeleteRegistration(ctx context.Context, gameID int64, registrationID int64) error {
	err := u.repository.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
		txErr := u.repository.LockGame(ctx, tx, gameID)
		if txErr != nil {
			return txErr
		}

		return u.repository.DeleteRegistration(ctx, tx, gameID, registrationID, u.clock.Now())
	})
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}

	return err
}
