package usecase

import (
	"context"

	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/utils"
)

type ListRegistrationsRepository interface {
	ListRegistrations(ctx context.Context, gameID int64) ([]model.Registrations, error)
}

// Registration - регистрация и её порядковый номер на игре
type Registration struct {
	model.Registrations

	Number int64
}

// ListRegistrations - регистрации на игру в порядке подачи, время по Москве
func (u usecase) ListRegistrations(ctx context.Context, gameID int64) ([]Registration, error) {
	res, err := u.repository.ListRegistrations(ctx, gameID)
	if err != nil {
		return nil, err
	}

	ret := make([]Registration, 0, len(res))
	for i, v := range res {
		v.CreatedAt = v.CreatedAt.In(utils.LocMsk)
		ret = append(ret, Registration{
			Registrations: v,
			Number:        int64(i + 1),
		})
	}

	return ret, nil
}
