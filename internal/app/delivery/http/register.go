package http

import (
	"context"
	"net/http"

	httpModel "quizon/internal/app/delivery/http/model"
)

type RegisterUsecase interface {
	Register(ctx context.Context, req httpModel.Register) error
}

func (d *delivery) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req httpModel.Register
	err := UnmarshalRequest(r.Body, &req)
	if err != nil {
		ResponseWithJSON(w, http.StatusBadRequest, Error{Msg: err.Error()})
		return
	}

	err = d.usecase.Register(ctx, req)
	if err != nil {
		ResponseWithJSON(w, http.StatusInternalServerError, Error{Msg: err.Error()})
		return
	}

	ResponseWithJSON(w, http.StatusOK, nil)
}
