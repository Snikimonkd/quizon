package http

import (
	"context"
	"net/http"

	httpModel "quizon/internal/app/delivery/http/model"
	"quizon/internal/pkg/logger"
)

type RegisterAvailableUsecase interface {
	RegisterAvailable(ctx context.Context, gameID int64) (httpModel.RegistrationStatus, error)
}

func (d *delivery) RegisterAvailable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req httpModel.RegisterAvailableRequest
	err := UnmarshalRequest(r.Body, &req)
	if err != nil {
		ResponseWithJSON(w, http.StatusBadRequest, Error{Msg: err.Error()})
		return
	}

	status, err := d.usecase.RegisterAvailable(ctx, req.GameID)
	if err != nil {
		logger.Errorf(err.Error())
		ResponseWithJSON(w, http.StatusInternalServerError, Error{Msg: err.Error()})
		return
	}

	ResponseWithJSON(w, http.StatusOK, httpModel.RegisterAvailableResponse{Available: status})
}
