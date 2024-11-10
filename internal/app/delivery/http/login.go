package http

import (
	"context"
	"errors"
	"net/http"

	httpModel "quizon/internal/app/delivery/http/model"
	"quizon/internal/app/usecase"
)

const authorizationTokenName string = `authorization-token`

type LoginUsecase interface {
	Login(ctx context.Context, req httpModel.LoginRequest) (usecase.Cookie, error)
}

func (d *delivery) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req httpModel.LoginRequest
	err := UnmarshalRequest(r.Body, &req)
	if err != nil {
		ResponseWithJSON(w, http.StatusBadRequest, Error{Msg: err.Error()})
		return
	}

	cookie, err := d.usecase.Login(ctx, req)
	if errors.Is(err, usecase.ErrWrongPassword) {
		ResponseWithJSON(w, http.StatusBadRequest, Error{Msg: err.Error()})
		return
	}
	if err != nil {
		ResponseWithJSON(w, http.StatusInternalServerError, Error{Msg: err.Error()})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name: authorizationTokenName,
		Path: "/",
		// Domain: "localhost:8000",
		// https
		Secure: false,
		// only visible to browser and not to js
		HttpOnly: true,
		// send from any domain to backend
		SameSite: http.SameSiteLaxMode,

		Value:   cookie.Value,
		Expires: cookie.ExpiresAt,
	})

	ResponseWithJSON(w, http.StatusOK, nil)
}
