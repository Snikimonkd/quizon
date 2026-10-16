package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"quizon/internal/generated/postgres/public/model"
)

var ErrNotOpenedYet error = errors.New("registration is not openned yet")

type RegisterRepository interface {
	Transactional(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
	LockGame(ctx context.Context, tx pgx.Tx, gameID int64) error
	GetRegistrationsAmount(ctx context.Context, tx pgx.Tx, gameID int64) (int64, int64, time.Time, int64, error)
	CreateRegistration(ctx context.Context, tx pgx.Tx, in model.Registrations) error
}

// CreateRegistration - регистрирует команду на игру. Если мест нет, возвращает StatusClosed и ничего не сохраняет.
func (u usecase) CreateRegistration(ctx context.Context, in model.Registrations) (RegistrationStatus, error) {
	now := u.clock.Now()
	in.CreatedAt = now

	var status RegistrationStatus

	err := u.repository.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
		txErr := u.repository.LockGame(ctx, tx, in.GameID)
		if txErr != nil {
			return txErr
		}

		amount, reserve, registrationOpenTime, cnt, txErr := u.repository.GetRegistrationsAmount(ctx, tx, in.GameID)
		if txErr != nil {
			return txErr
		}

		if now.Before(registrationOpenTime) {
			return ErrNotOpenedYet
		}

		status = registrationStatus(cnt, amount, reserve)
		if status == StatusClosed {
			return nil
		}

		return u.repository.CreateRegistration(ctx, tx, in)
	})
	if err != nil {
		return "", err
	}

	return status, nil
}
