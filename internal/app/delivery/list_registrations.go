package delivery

import (
	"context"

	"quizon/internal/app/delivery/api"
)

type ListRegistrationsUsecase interface {
	ListRegistrations(ctx context.Context, gameID int64) (api.GetGamesIdRegistrationsResponseObject, error)
}

func (d delivery) GetGamesIdRegistrations(
	ctx context.Context,
	req api.GetGamesIdRegistrationsRequestObject,
) (api.GetGamesIdRegistrationsResponseObject, error) {
	resp, err := d.usecase.ListRegistrations(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return resp, err
}
