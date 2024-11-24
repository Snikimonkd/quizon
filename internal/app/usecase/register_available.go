package usecase

// import (
// 	"context"
// 	"time"
//
// 	httpModel "quizon/internal/app/delivery/model"
// 	"quizon/internal/generated/postgres/public/model"
// 	"quizon/internal/utils"
// )
//
// type RegisterAvailableRepository interface {
// 	GetGame(ctx context.Context, gameID int64) (model.Games, error)
// 	RegistrationsAmount(ctx context.Context, gameID int64) (int64, error)
// }
//
// func (u usecase) RegisterAvailable(ctx context.Context, gameID int64) (httpModel.RegistrationStatus, error) {
// 	restrictionsLimitations, err := u.repository.GetGame(
// 		ctx,
// 		gameID,
// 	)
// 	if err != nil {
// 		return httpModel.RegistrationStatus(""), err
// 	}
//
// 	if !time.Now().In(utils.LocMsk).After(restrictionsLimitations.StartTime.In(utils.LocMsk)) {
// 		return httpModel.NotOpenedYet, nil
// 	}
//
// 	regsAmount, err := u.repository.RegistrationsAmount(ctx, gameID)
// 	if err != nil {
// 		return httpModel.RegistrationStatus(""), err
// 	}
//
// 	if regsAmount < restrictionsLimitations.MainAmount {
// 		return httpModel.Available, nil
// 	}
//
// 	if regsAmount < restrictionsLimitations.MainAmount+restrictionsLimitations.ReserveAmount {
// 		return httpModel.Reserve, nil
// 	}
//
// 	return httpModel.Closed, nil
// }
