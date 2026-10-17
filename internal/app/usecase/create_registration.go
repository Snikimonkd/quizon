package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"quizon/internal/generated/postgres/public/model"
)

var ErrNotOpenedYet error = errors.New("registration is not openned yet")

type RegisterRepository interface {
	// GetGameWithRegistrations - игра и число оставшихся в ней команд в транзакции tx. ErrNotFound, если игры нет
	GetGameWithRegistrations(ctx context.Context, tx pgx.Tx, gameID int64) (GameWithRegistrations, error)
	CreateRegistration(ctx context.Context, tx pgx.Tx, in model.Registrations) error
}

// CreateRegistration - регистрирует команду на игру, пока регистрация открыта и есть места.
// Иначе ничего не сохраняет и возвращает статус-причину: StatusClosed — мест нет, StatusFinished — игра уже началась.
// ErrNotOpenedYet, если регистрация ещё не открылась
func (u usecase) CreateRegistration(ctx context.Context, in model.Registrations) (RegistrationStatus, error) {
	now := u.clock.Now()
	in.CreatedAt = now

	var status RegistrationStatus
	err := u.repository.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// пока считаем места, их никто не займёт и не освободит
		txErr := u.repository.LockGame(ctx, tx, in.GameID)
		if txErr != nil {
			return txErr
		}

		game, txErr := u.repository.GetGameWithRegistrations(ctx, tx, in.GameID)
		if txErr != nil {
			return txErr
		}

		// тот же статус, что капитаны видят в списке игр, но посчитанный под блокировкой игры
		status = gameStatus(game.Games, game.RegistrationsAmount, now)
		switch status {
		case StatusOk, StatusReserve:
			return u.repository.CreateRegistration(ctx, tx, in)
		case StatusWaiting:
			return ErrNotOpenedYet
		case StatusClosed, StatusFinished:
		}

		// мест нет или игра уже началась: ничего не сохраняем
		return nil
	})
	if err != nil {
		return "", err
	}

	return status, nil
}
