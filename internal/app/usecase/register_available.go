package usecase

import (
	"context"
	"time"

	httpModel "quizon_bot/internal/app/delivery/http/model"
	"quizon_bot/internal/generated/postgres/public/model"
	"quizon_bot/internal/utils"
)

type RegisterAvailableRepository interface {
	SelectRegistrationRestrictions(ctx context.Context) (model.Games, error)
	RegistrationsAmount(ctx context.Context) (int64, error)
}

func (u usecase) RegisterAvailable(ctx context.Context) (httpModel.RegistrationStatus, error) {
	restrictionsLimitations, err := u.registerAvailableRepository.SelectRegistrationRestrictions(
		ctx,
	)
	if err != nil {
		return httpModel.RegistrationStatus(""), err
	}

	if !time.Now().In(utils.LocMsk).After(restrictionsLimitations.OpenningTime.In(utils.LocMsk)) {
		return httpModel.NotOpenedYet, nil
	}

	regsAMount, err := u.registerAvailableRepository.RegistrationsAmount(ctx)
	if err != nil {
		return httpModel.RegistrationStatus(""), err
	}

	if regsAMount < restrictionsLimitations.Reserve {
		return httpModel.Available, nil
	}

	if regsAMount < restrictionsLimitations.Closed {
		return httpModel.Reserve, nil
	}

	return httpModel.Closed, nil
}
