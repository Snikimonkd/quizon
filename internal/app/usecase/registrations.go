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

	// Number - место в очереди среди оставшихся в игре команд, у удалённых 0
	Number int64
}

// ListRegistrations - все регистрации на игру в порядке подачи, включая удалённые; время по Москве
func (u usecase) ListRegistrations(ctx context.Context, gameID int64) ([]Registration, error) {
	res, err := u.repository.ListRegistrations(ctx, gameID)
	if err != nil {
		return nil, err
	}

	ret := make([]Registration, 0, len(res))
	number := int64(1)
	for _, v := range res {
		v.CreatedAt = v.CreatedAt.In(utils.LocMsk)
		reg := Registration{Registrations: v}
		if reg.DeletedAt == nil {
			reg.Number = number
			number++
		}
		ret = append(ret, reg)
	}

	return ret, nil
}
