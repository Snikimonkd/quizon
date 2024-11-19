package delivery

import (
	"context"

	"quizon/internal/app/delivery/api"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ListGamesUsecase interface {
	ListGames(ctx context.Context, page int64, perPage int64) (api.GetGamesResponseObject, error)
}

func (d delivery) GetGames(ctx context.Context, req api.GetGamesRequestObject) (api.GetGamesResponseObject, error) {
	err := validateListGamesRequest(req.Params)
	if err != nil {
		return api.GetGames400JSONResponse{
			Error: err.Error(),
		}, nil
	}

	resp, err := d.usecase.ListGames(ctx, req.Params.Page, req.Params.PerPage)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func validateListGamesRequest(req api.GetGamesParams) error {
	return validation.ValidateStruct(
		&req,
		validation.Field(&req.Page, validation.Required, validation.Min(1)),
		validation.Field(&req.PerPage, validation.Required, validation.Min(1), validation.Max(100)),
	)
}
