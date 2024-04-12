package usecase

import (
	"context"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
)

type ListGamesRepository interface {
	ListGames(ctx context.Context, limit int64, offset int64) ([]model.Game, error)
}

func (u usecase) ListGames(ctx context.Context, limit int64, offset int64) ([]model.Game, error) {
	return u.listGamesRepository.ListGames(ctx, limit, offset)
}
