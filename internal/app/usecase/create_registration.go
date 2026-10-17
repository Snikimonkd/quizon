package usecase

import (
	"context"
	"errors"
	"unicode"

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
// ValidationError, если анкета нарушает правила, ErrNotOpenedYet — если регистрация ещё не открылась
func (u usecase) CreateRegistration(ctx context.Context, in model.Registrations) (RegistrationStatus, error) {
	err := validateRegistration(in)
	if err != nil {
		return "", err
	}

	now := u.clock.Now()
	in.CreatedAt = now

	var status RegistrationStatus
	err = u.repository.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
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

// validateRegistration - правила анкеты: капитан, телефон, Telegram и название команды указаны,
// в телефоне от 10 до 15 цифр, придёт хотя бы один человек
func validateRegistration(r model.Registrations) error {
	invalid := ValidationError{}
	invalid.check(FieldCaptainName, filled(r.CaptainName), ErrRequired)
	invalid.check(FieldPhone, filled(r.Phone), ErrRequired)
	invalid.check(FieldPhone, validPhone(r.Phone), ErrBadPhone)
	invalid.check(FieldTelegram, filled(r.Telegram), ErrRequired)
	invalid.check(FieldTeamName, filled(r.TeamName), ErrRequired)
	invalid.check(FieldPlayersAmount, r.PlayersAmount >= 1, ErrTooSmall)

	return invalid.err()
}

// validPhone - в номере от 10 до 15 цифр, остальное (+, скобки, пробелы, дефисы) не важно
func validPhone(phone string) bool {
	digits := 0
	for _, r := range phone {
		if unicode.IsDigit(r) {
			digits++
		}
	}

	return digits >= 10 && digits <= 15
}
