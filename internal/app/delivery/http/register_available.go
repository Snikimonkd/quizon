package http

import (
	"context"
	"net/http"

	httpModel "quizon_bot/internal/app/delivery/http/model"
	"quizon_bot/internal/pkg/logger"
)

type RegisterAvailableUsecase interface {
	RegisterAvailable(ctx context.Context) (httpModel.RegistrationStatus, error)
}

func (d *delivery) RegisterAvailable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status, err := d.usecase.RegisterAvailable(ctx)
	if err != nil {
		logger.Errorf(err.Error())
		ResponseWithJSON(w, http.StatusInternalServerError, Error{Msg: err.Error()})
		return
	}

	ResponseWithJSON(w, http.StatusOK, httpModel.RegisterAvailable{Available: status})
}
