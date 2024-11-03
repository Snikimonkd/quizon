package usecase

import (
	"context"

	httpModel "quizon_bot/internal/app/delivery/http/model"
	"quizon_bot/internal/generated/postgres/public/model"
)

type RegisterRepository interface {
	Register(ctx context.Context, in model.Registrations) error
}

func (u usecase) Register(ctx context.Context, req httpModel.Register) error {
	now := u.clock.Now()
	domainModel := model.Registrations{
		TgContact:   req.TgContact,
		TeamID:      req.TeamID,
		TeamName:    req.TeamName,
		CaptainName: req.CaptainName,
		Phone:       req.Phone,
		GroupName:   req.GroupName,
		Amount:      req.Amount,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := u.registerRepository.Register(ctx, domainModel)
	if err != nil {
		return err
	}

	return nil
}
