package usecase

import (
	"context"

	"quizon/internal/generated/postgres/public/model"
)

type CreateGameRepository interface {
	CreateGame(ctx context.Context, in model.Games) (int64, error)
}

// CreateGame - создаёт игру и возвращает её id. ValidationError, если игра нарушает правила
func (u usecase) CreateGame(ctx context.Context, in model.Games) (int64, error) {
	err := validateGame(in)
	if err != nil {
		return 0, err
	}

	in.CreatedAt = u.clock.Now()

	return u.repository.CreateGame(ctx, in)
}

// validateGame - правила игры: название и место указаны, время начала и открытия регистрации заданы,
// регистрация открывается раньше начала, в основе хотя бы одно место, резерв не отрицательный
func validateGame(g model.Games) error {
	invalid := ValidationError{}
	invalid.check(FieldName, filled(g.Name), ErrRequired)
	invalid.check(FieldLocation, filled(g.Location), ErrRequired)
	invalid.check(FieldStartTime, !g.StartTime.IsZero(), ErrRequired)
	invalid.check(FieldRegistrationOpenTime, !g.RegistrationOpenTime.IsZero(), ErrRequired)
	// без времени начала порядок не проверить, об этом уже скажет StartTime
	ordered := g.StartTime.IsZero() || g.RegistrationOpenTime.Before(g.StartTime)
	invalid.check(FieldRegistrationOpenTime, ordered, ErrOpensAfterStart)
	invalid.check(FieldMainAmount, g.MainAmount >= 1, ErrTooSmall)
	invalid.check(FieldReserveAmount, g.ReserveAmount >= 0, ErrTooSmall)

	return invalid.err()
}
