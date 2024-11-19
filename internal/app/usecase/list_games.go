package usecase

import (
	"context"

	"quizon/internal/app/delivery/api"
	"quizon/internal/app/repository"
)

type ListGamesRepository interface {
	ListGames(ctx context.Context, page int64, perPage int64) ([]repository.GameWithRegistrations, error)
}

func (u usecase) ListGames(ctx context.Context, page int64, perPage int64) (api.GetGamesResponseObject, error) {
	games, err := u.repository.ListGames(ctx, page, perPage)
	if err != nil {
		return nil, err
	}

	ret := make([]api.ListGamesItem, 0, len(games))
	for _, v := range games {
		var status api.RegistrationStatus
		if v.MainAmount+v.ReserveAmount <= v.RegistrationsAmount {
			status = api.Closed
		}
		if v.MainAmount <= v.RegistrationsAmount && v.RegistrationsAmount <= v.MainAmount+v.ReserveAmount {
			status = api.Reserve
		}
		if v.RegistrationsAmount < v.MainAmount {
			status = api.Ok
		}

		ret = append(ret, api.ListGamesItem{
			Id:                   v.ID,
			StartTime:            v.StartTime,
			Location:             v.Location,
			Name:                 v.Name,
			MainAmount:           v.MainAmount,
			ReserveAmount:        v.ReserveAmount,
			RegistrationOpenTime: v.RegistrationOpenTime,
			RegistrationStatus:   status,
		})
	}

	return api.GetGames200JSONResponse(ret), nil
}
