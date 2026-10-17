package usecase

import (
	"context"

	"quizon/internal/generated/postgres/public/model"
)

type GetGameRepository interface {
	// GetGame - игра по id. ErrNotFound, если такой нет
	GetGame(ctx context.Context, gameID int64) (model.Games, error)
	RegistrationsAmount(ctx context.Context, gameID int64) (int64, error)
}

// GetGame - получить игру по id и статус регистрации на неё сейчас, как в списке игр. ErrNotFound, если игры нет
func (u usecase) GetGame(ctx context.Context, gameID int64) (GameListItem, error) {
	game, err := u.repository.GetGame(ctx, gameID)
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
