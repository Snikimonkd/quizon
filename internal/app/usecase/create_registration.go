package usecase

import (
	"context"
	"errors"
	"time"

	"quizon/internal/app/delivery/api"
	"quizon/internal/generated/postgres/public/model"

	"github.com/jackc/pgx/v5"
)

var ErrNotOpenedYet error = errors.New("registration is not openned yet")

type RegisterRepository interface {
	Transactional(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
	LockGame(ctx context.Context, tx pgx.Tx, gameID int64) error
	GetRegistrationsAmount(ctx context.Context, tx pgx.Tx, gameID int64) (int64, int64, time.Time, int64, error)
	CreateRegistration(ctx context.Context, tx pgx.Tx, in model.Registrations) error
}

func (u usecase) CreateRegistration(
	ctx context.Context,
	req api.PostRegistrationRequestObject,
) (api.PostRegistrationResponseObject, error) {
	now := u.clock.Now()
	dbModel := model.Registrations{
		GameID:        req.Body.GameId,
		CreatedAt:     now,
		TeamName:      req.Body.TeamName,
		CaptainName:   req.Body.CaptainName,
		Phone:         req.Body.Phone,
		Telegram:      req.Body.Telegram,
		PlayersAmount: req.Body.PlayersAmount,
		GroupName:     req.Body.GroupName,
		TeamID:        req.Body.TeamId,
	}

	var status api.RegistrationStatus

	err := u.repository.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
		txErr := u.repository.LockGame(ctx, tx, req.Body.GameId)
		if txErr != nil {
			return txErr
		}

		amount, reserve, registrationOpenTime, cnt, txErr := u.repository.GetRegistrationsAmount(
			ctx,
			tx,
			req.Body.GameId,
		)
		if txErr != nil {
			return txErr
		}

		if time.Now().Before(registrationOpenTime) {
			return ErrNotOpenedYet
		}

		if amount+reserve <= cnt {
			status = api.Closed
			return nil
		}

		if cnt >= amount && cnt <= amount+reserve {
			status = api.Reserve
		}
		if cnt < amount {
			status = api.Ok
		}

		txErr = u.repository.CreateRegistration(ctx, tx, dbModel)
		if txErr != nil {
			return txErr
		}

		return nil
	})
	if errors.Is(err, ErrNotOpenedYet) {
		return api.PostRegistration400JSONResponse{
			Error: err.Error(),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return api.PostRegistration200JSONResponse{
		Status: status,
	}, nil
}
