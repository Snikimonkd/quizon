package usecase

import (
	"context"

	httpModel "quizon/internal/app/delivery/model"
	"quizon/internal/generated/postgres/public/model"
)

type ListGamesRepository interface {
	ListGames(ctx context.Context) ([]model.Games, error)
}

func (u usecase) ListGames(ctx context.Context) ([]httpModel.Game, error) {
	games, err := u.repository.ListGames(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]httpModel.Game, 0, len(games))
	for _, v := range games {
		ret = append(ret, httpModel.Game{
			ID:                   v.ID,
			CreatedAt:            v.CreatedAt,
			StartTime:            v.StartTime,
			Location:             v.Location,
			Name:                 v.Name,
			MainAmount:           v.MainAmount,
			ReserveAmount:        v.ReserveAmount,
			RegistartionOpenTime: v.RegistartionOpenTime,
		})
	}

	return ret, nil
}
