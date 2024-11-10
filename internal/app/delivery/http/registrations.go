package http

import (
	"context"
	"net/http"
	"strconv"

	httpModel "quizon/internal/app/delivery/http/model"
)

type RegistrationsUsecase interface {
	Registrations(ctx context.Context, gameID int64) ([]httpModel.Registration, error)
}

func (d *delivery) Registrations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	gameIDStr := r.URL.Query().Get("game_id")
	gameID, err := strconv.ParseInt(gameIDStr, 10, 64)
	if err != nil {
		ResponseWithJSON(w, http.StatusBadRequest, Error{Msg: "can't parse game id: " + err.Error()})
		return
	}

	res, err := d.usecase.Registrations(ctx, gameID)
	if err != nil {
		ResponseWithJSON(w, http.StatusInternalServerError, Error{Msg: err.Error()})
		return
	}

	ResponseWithJSON(w, http.StatusOK, res)
}
