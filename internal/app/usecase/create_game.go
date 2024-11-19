package usecase

import (
	"context"

	"quizon/internal/app/delivery/api"
	"quizon/internal/generated/postgres/public/model"
)

type CreateGameRepository interface {
	CreateGame(ctx context.Context, in model.Games) (int64, error)
}

func (u usecase) CreateGame(ctx context.Context, req api.PostGameRequestObject) (api.PostGameResponseObject, error) {
	now := u.clock.Now()
	domainModel := model.Games{
		CreatedAt:            now,
		StartTime:            req.Body.StartTime,
		Location:             req.Body.Location,
		Name:                 req.Body.Name,
		MainAmount:           req.Body.MainAmount,
		ReserveAmount:        req.Body.ReserveAmount,
		RegistrationOpenTime: req.Body.RegistrationOpenTime,
	}

	gameID, err := u.repository.CreateGame(ctx, domainModel)
	if err != nil {
		return nil, err
	}

	return api.PostGame200JSONResponse{
		Id: gameID,
	}, nil
}
