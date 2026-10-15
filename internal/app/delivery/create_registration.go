package delivery

import (
	"context"

	"quizon/internal/app/delivery/api"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type CreateRegistrationUsecase interface {
	CreateRegistration(
		ctx context.Context,
		req api.PostRegistrationRequestObject,
	) (api.PostRegistrationResponseObject, error)
}

func (d delivery) PostRegistration(
	ctx context.Context,
	req api.PostRegistrationRequestObject,
) (api.PostRegistrationResponseObject, error) {
	err := validateCreateRegistrationRequest(req.Body)
	if err != nil {
		return api.PostRegistration400JSONResponse{
			Error: err.Error(),
		}, nil
	}

	resp, err := d.usecase.CreateRegistration(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func validateCreateRegistrationRequest(req *api.CreateRegistrationRequest) error {
	return validation.ValidateStruct(
		req,
		validation.Field(&req.CaptainName, validation.Required),
		validation.Field(&req.GameId, validation.Required),
		validation.Field(&req.Phone, validation.Required),
		validation.Field(&req.PlayersAmount, validation.Required),
		validation.Field(&req.TeamName, validation.Required),
		validation.Field(&req.Telegram, validation.Required),
	)
}
