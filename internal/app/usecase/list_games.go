package usecase

import (
	"context"

	"quizon/internal/generated/postgres/public/model"
)

type ListGamesRepository interface {
	ListGames(ctx context.Context, page int64, perPage int64) ([]GameWithRegistrations, error)
}

// GameWithRegistrations - игра и число оставшихся в ней команд, удалённые не считаются
type GameWithRegistrations struct {
	model.Games

	RegistrationsAmount int64
}

// GameListItem - игра и статус регистрации на неё в момент запроса
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

	// один момент на весь список: игры не должны расходиться во времени
	now := u.clock.Now()
	ret := make([]GameListItem, 0, len(games))
	for _, v := range games {
		ret = append(ret, GameListItem{
			Games:  v.Games,
			Status: gameStatus(v.Games, v.RegistrationsAmount, now),
		})
	}

	return ret, nil
}
