package usecase

import (
	"context"

	httpModel "quizon/internal/app/delivery/model"
	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/utils"
)

type RegistrationsRepository interface {
	Registrations(ctx context.Context, gameID int64) ([]model.Registrations, error)
}

func (u usecase) Registrations(ctx context.Context, gameID int64) ([]httpModel.Registration, error) {
	res, err := u.repository.Registrations(ctx, gameID)
	if err != nil {
		return nil, err
	}

	ret := make([]httpModel.Registration, 0, len(res))
	for i, v := range res {
		ret = append(ret,
			httpModel.Registration{
				Number:        int64(i + 1),
				Telegram:      v.Telegram,
				TeamID:        v.TeamID,
				TeamName:      v.TeamName,
				CaptainName:   v.CaptainName,
				Phone:         v.Phone,
				GroupName:     v.GroupName,
				PlayersAmount: v.PlayersAmount,
				RegisteredAt:  utils.PrettyTime(v.CreatedAt),
			},
		)
	}

	return ret, nil
}
