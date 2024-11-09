package usecase

import (
	"context"

	httpModel "quizon/internal/app/delivery/http/model"
	"quizon/internal/generated/postgres/public/model"
)

type RegisterRepository interface {
	Register(ctx context.Context, in model.Registrations) error
}

func (u usecase) Register(ctx context.Context, req httpModel.Register) error {
	now := u.clock.Now()
	domainModel := model.Registrations{
		GameID:      req.GameID,
		CreatedAt:   now,
		TeamName:    req.TeamName,
		CaptainName: req.CaptainName,
		Phone:       req.Phone,
		Telegram:    req.TgContact,
		TeamSize:    req.TeamSize,
		GroupName:   req.GroupName,
		TeamID:      req.TeamID,
	}

	err := u.repository.Register(ctx, domainModel)
	if err != nil {
		return err
	}

	return nil
}
