package usecase

import (
	"context"

	"quizon/internal/generated/postgres/public/model"
)

type ListRegistrationsRepository interface {
	ListRegistrations(ctx context.Context, gameID int64) ([]model.Registrations, error)
}

// Registration - регистрация, её номер и место в игре
type Registration struct {
	model.Registrations

	// Number - место в очереди среди оставшихся в игре команд, у удалённых 0
	Number int64
	// Deleted - команду убрали из игры: она остаётся в списке, но места не занимает
	Deleted bool
	// Reserve - команда в резерве: её номер больше мест в основе
	Reserve bool
	// FirstReserve - первая команда резерва: займёт место команды, которую уберут из основы
	FirstReserve bool
}

// ListRegistrations - все регистрации на игру в порядке подачи, включая удалённые, с номером и местом в игре.
// ErrNotFound, если игры нет
func (u usecase) ListRegistrations(ctx context.Context, gameID int64) ([]Registration, error) {
	game, err := u.repository.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}

	res, err := u.repository.ListRegistrations(ctx, gameID)
	if err != nil {
		return nil, err
	}

	ret := make([]Registration, 0, len(res))
	number := int64(1)
	for _, v := range res {
		reg := Registration{Registrations: v, Deleted: v.DeletedAt != nil}
		if !reg.Deleted {
			reg.Number = number
			reg.Reserve = number > game.MainAmount
			reg.FirstReserve = number == game.MainAmount+1
			number++
		}
		ret = append(ret, reg)
	}

	return ret, nil
}
