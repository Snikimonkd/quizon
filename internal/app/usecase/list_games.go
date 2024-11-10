package usecase

import (
	"context"

	"quizon/internal/generated/postgres/public/model"
)

type ListGamesRepository interface {
	ListGames(ctx context.Context) ([]model.Games, error)
}

func (u usecase) ListGames(ctx context.Context) ([]model.Games, error) {
	games, err := u.repository.ListGames(ctx)
	if err != nil {
		return nil, err
	}

	return games, nil
}
