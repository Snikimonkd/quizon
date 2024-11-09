package usecase

import (
	"context"

	httpModel "quizon/internal/app/delivery/http/model"
	"quizon/internal/generated/postgres/public/model"
)

type CreateGameRepository interface {
	CreateGame(ctx context.Context, in model.Games) error
}

func (u usecase) CreateGame(ctx context.Context, req httpModel.CreateGameRequest) error {
	now := u.clock.Now()
	domainModel := model.Games{
		CreatedAt:            now,
		StartTime:            req.StartTime,
		Location:             req.Location,
		Name:                 req.Name,
		MainAmount:           req.MainAmount,
		ReserveAmount:        req.ReserveAmount,
		RegistartionOpenTime: req.RegistartionOpenTime,
	}

	err := u.repository.CreateGame(ctx, domainModel)
	if err != nil {
		return err
	}

	return nil
}
