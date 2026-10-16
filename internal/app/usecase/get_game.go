package usecase

import (
	"context"
	"errors"

	"quizon/internal/generated/postgres/public/model"

	"github.com/jackc/pgx/v5"
)

type GetGameRepository interface {
	GetGame(ctx context.Context, gameID int64) (model.Games, error)
}

// GetGame - получить игру по id
func (u usecase) GetGame(ctx context.Context, gameID int64) (model.Games, error) {
	game, err := u.repository.GetGame(ctx, gameID)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Games{}, ErrNotFound
	}
	if err != nil {
		return model.Games{}, err
	}

	return game, nil
}
