package usecase

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type DeleteRegistrationRepository interface {
	// DeleteRegistration - помечает регистрацию удалённой в транзакции tx.
	// ErrNotFound, если её нет в этой игре или она уже удалена
	DeleteRegistration(ctx context.Context, tx pgx.Tx, gameID int64, registrationID int64, deletedAt time.Time) error
}

// DeleteRegistration - убирает команду из игры. Под той же блокировкой игры, что и регистрация,
// чтобы новая команда не посчитала места, пока освобождается чужое.
// ErrNotFound, если команды нет в этой игре или она уже удалена
func (u usecase) DeleteRegistration(ctx context.Context, gameID int64, registrationID int64) error {
	return u.repository.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
		txErr := u.repository.LockGame(ctx, tx, gameID)
		if txErr != nil {
			return txErr
		}

		return u.repository.DeleteRegistration(ctx, tx, gameID, registrationID, u.clock.Now())
	})
}
