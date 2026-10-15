package delivery

import (
	"context"

	"quizon/internal/app/delivery/api"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type CreateGameUsecase interface {
	CreateGame(
		ctx context.Context,
		req api.PostGameRequestObject,
	) (api.PostGameResponseObject, error)
}

func (d delivery) PostGame(
	ctx context.Context,
	req api.PostGameRequestObject,
) (api.PostGameResponseObject, error) {
	err := validateCreateGameRequest(req.Body)
	if err != nil {
		return api.PostGame400JSONResponse{
			Error: err.Error(),
		}, nil
	}

	resp, err := d.usecase.CreateGame(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func validateCreateGameRequest(req *api.CreateGameRequest) error {
	return validation.ValidateStruct(
		req,
		validation.Field(&req.Location, validation.Required),
		validation.Field(&req.MainAmount, validation.Required),
		validation.Field(&req.Name, validation.Required),
		validation.Field(&req.RegistrationOpenTime, validation.Required),
		validation.Field(&req.ReserveAmount, validation.Required),
		validation.Field(&req.StartTime, validation.Required),
	)
}
