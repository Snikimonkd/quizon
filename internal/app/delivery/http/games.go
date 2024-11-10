package http

import (
	"context"
	"net/http"

	httpModel "quizon/internal/app/delivery/http/model"
	"quizon/internal/generated/postgres/public/model"
)

type ListGamesUsecase interface {
	ListGames(ctx context.Context) ([]model.Games, error)
}

func (d *delivery) Games(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	games, err := d.usecase.ListGames(ctx)
	if err != nil {
		ResponseWithJSON(w, http.StatusInternalServerError, Error{Msg: err.Error()})
		return
	}

	ret := make([]httpModel.Game, 0, len(games))
	for _, v := range games {
		ret = append(ret, httpModel.Game{
			ID:                   v.ID,
			CreatedAt:            v.CreatedAt,
			StartTime:            v.StartTime,
			Location:             v.Location,
			Name:                 v.Name,
			MainAmount:           v.MainAmount,
			ReserveAmount:        v.ReserveAmount,
			RegistartionOpenTime: v.RegistartionOpenTime,
		})
	}

	ResponseWithJSON(w, http.StatusOK, ret)
}
