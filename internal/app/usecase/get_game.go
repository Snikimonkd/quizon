package usecase

import (
	"context"
	"errors"

	"quizon/internal/generated/postgres/public/model"

	"github.com/jackc/pgx/v5"
)

type GetGameRepository interface {
	GetGame(ctx context.Context, gameID int64) (model.Games, error)
	RegistrationsAmount(ctx context.Context, gameID int64) (int64, error)
}

// GetGame - получить игру по id и статус регистрации на неё сейчас, как в списке игр
func (u usecase) GetGame(ctx context.Context, gameID int64) (GameListItem, error) {
	game, err := u.repository.GetGame(ctx, gameID)
	if errors.Is(err, pgx.ErrNoRows) {
		return GameListItem{}, ErrNotFound
	}
	if err != nil {
		return GameListItem{}, err
	}

	registered, err := u.repository.RegistrationsAmount(ctx, gameID)
	if err != nil {
		return GameListItem{}, err
	}

	return GameListItem{
		Games:  game,
		Status: gameStatus(game, registered, u.clock.Now()),
	}, nil
}
