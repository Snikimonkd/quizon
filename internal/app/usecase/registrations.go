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

// Deleted - команду убрали из игры
func (r Registration) Deleted() bool {
	return r.DeletedAt != nil
}

// ListRegistrations - все регистрации на игру в порядке подачи, включая удалённые; время по Москве
func (u usecase) ListRegistrations(ctx context.Context, gameID int64) ([]Registration, error) {
	res, err := u.repository.ListRegistrations(ctx, gameID)
	if err != nil {
		return nil, err
	}

	return numberRegistrations(res), nil
}

// numberRegistrations - нумерует команды по порядку подачи, удалённые пропускает
func numberRegistrations(in []model.Registrations) []Registration {
	ret := make([]Registration, 0, len(in))
	var number int64
	for _, v := range in {
		v.CreatedAt = v.CreatedAt.In(utils.LocMsk)
		reg := Registration{Registrations: v}
		if !reg.Deleted() {
			number++
			reg.Number = number
		}
		ret = append(ret, reg)
	}

	return ret
}
