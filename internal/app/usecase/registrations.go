package usecase

import (
	"context"

	"quizon/internal/app/delivery/api"
	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/utils"
)

type ListRegistrationsRepository interface {
	ListRegistrations(ctx context.Context, gameID int64) ([]model.Registrations, error)
}

func (u usecase) ListRegistrations(
	ctx context.Context,
	gameID int64,
) (api.GetGamesIdRegistrationsResponseObject, error) {
	res, err := u.repository.ListRegistrations(ctx, gameID)
	if err != nil {
		return nil, err
	}

	ret := make([]api.ListRegistrationsItem, 0, len(res))
	for i, v := range res {
		ret = append(ret,
			api.ListRegistrationsItem{
				Number:        int64(i + 1),
				Telegram:      v.Telegram,
				TeamId:        v.TeamID,
				TeamName:      v.TeamName,
				CaptainName:   v.CaptainName,
				Phone:         v.Phone,
				GroupName:     v.GroupName,
				PlayersAmount: v.PlayersAmount,
				RegisteredAt:  v.CreatedAt.In(utils.LocMsk),
			},
		)
	}

	return api.GetGamesIdRegistrations200JSONResponse(ret), nil
}
