package usecase

import (
	"context"

	"quizon/internal/app/repository"
	"quizon/internal/generated/postgres/public/model"
)

type ListGamesRepository interface {
	ListGames(ctx context.Context, page int64, perPage int64) ([]repository.GameWithRegistrations, error)
}

// GameListItem - игра и статус регистрации на неё
type GameListItem struct {
	model.Games

	Status RegistrationStatus
}

// ListGames - список игр, новые сверху
func (u usecase) ListGames(ctx context.Context, page int64, perPage int64) ([]GameListItem, error) {
	games, err := u.repository.ListGames(ctx, page, perPage)
	if err != nil {
		return nil, err
	}

	ret := make([]GameListItem, 0, len(games))
	for _, v := range games {
		ret = append(ret, GameListItem{
			Games:  v.Games,
			Status: registrationStatus(v.RegistrationsAmount, v.MainAmount, v.ReserveAmount),
		})
	}

	return ret, nil
}
