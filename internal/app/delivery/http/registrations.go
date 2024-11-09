package http

import (
	"context"
	"net/http"

	httpModel "quizon/internal/app/delivery/http/model"
)

type RegistrationsUsecase interface {
	Registrations(ctx context.Context, gameID int64) ([]httpModel.Registration, error)
}

func (d *delivery) Registrations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req httpModel.ListRegistrationsRequest
	err := UnmarshalRequest(r.Body, &req)
	if err != nil {
		ResponseWithJSON(w, http.StatusBadRequest, Error{Msg: err.Error()})
		return
	}

	//	if req.Password != "09154cb6-f723-4f3d-943c-7a6e4b155eb1" {
	//		ResponseWithJSON(w, http.StatusUnauthorized, Error{Msg: "ti po moemu chto-to pereputal"})
	//		return
	//	}

	res, err := d.usecase.Registrations(ctx, req.GameID)
	if err != nil {
		ResponseWithJSON(w, http.StatusInternalServerError, Error{Msg: err.Error()})
		return
	}

	ResponseWithJSON(w, http.StatusOK, res)
}
