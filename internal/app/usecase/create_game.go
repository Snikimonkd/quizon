package usecase

import (
	"context"
	"errors"

	"quizon/internal/generated/postgres/public/model"
)

// ErrOpensAfterStart - регистрация должна открыться раньше, чем начнётся игра
var ErrOpensAfterStart = errors.New("registration must open before the game starts")

type CreateGameRepository interface {
	CreateGame(ctx context.Context, in model.Games) (int64, error)
}

// CreateGame - создаёт игру и возвращает её id.
// ErrOpensAfterStart, если регистрация открывается не раньше начала игры
func (u usecase) CreateGame(ctx context.Context, in model.Games) (int64, error) {
	if !in.RegistrationOpenTime.Before(in.StartTime) {
		return 0, ErrOpensAfterStart
	}

	in.CreatedAt = u.clock.Now()

	return u.repository.CreateGame(ctx, in)
}
