package delivery

import (
	"context"
	"net/http"

	httpModel "quizon/internal/app/delivery/model"
)

type ListGamesUsecase interface {
	ListGames(ctx context.Context) ([]httpModel.Game, error)
}

func (d *delivery) Games(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	games, err := d.usecase.ListGames(ctx)
	if err != nil {
		ResponseWithJSON(w, http.StatusInternalServerError, Error{Msg: err.Error()})
		return
	}

	ResponseWithJSON(w, http.StatusOK, games)
}
