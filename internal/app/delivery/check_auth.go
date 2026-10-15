package delivery

import (
	"context"

	"quizon/internal/app/delivery/api"
)

func (d delivery) GetAuth(_ context.Context, _ api.GetAuthRequestObject) (api.GetAuthResponseObject, error) {
	return api.GetAuth200JSONResponse{}, nil
}
