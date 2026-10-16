package usecase

import (
	"context"

	"quizon/internal/generated/postgres/public/model"
)

type CreateGameRepository interface {
	CreateGame(ctx context.Context, in model.Games) (int64, error)
}

// CreateGame - создаёт игру и возвращает её id
func (u usecase) CreateGame(ctx context.Context, in model.Games) (int64, error) {
	in.CreatedAt = u.clock.Now()

	return u.repository.CreateGame(ctx, in)
}
