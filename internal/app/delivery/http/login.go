package http

import (
	"net/http"
	"time"

	httpModel "quizon/internal/app/delivery/http/model"
)

// type LoginUsecase interface {
// 	Login(ctx context.Context, req httpModel.Login) error
// }

func (d *delivery) Login(w http.ResponseWriter, r *http.Request) {
	_ = r.Context()
	var req httpModel.LoginRequest
	err := UnmarshalRequest(r.Body, &req)
	if err != nil {
		ResponseWithJSON(w, http.StatusBadRequest, Error{Msg: err.Error()})
		return
	}

	//	err = d.loginUsecase.Login(ctx, req)
	//	if err != nil {
	//		logger.Error(err.Error())
	//		ResponseWithJson(w, http.StatusInternalServerError, Error{Msg: err.Error()})
	//		return
	//	}

	http.SetCookie(w, &http.Cookie{
		Name:     "authorization",
		Value:    "0e689e44-9d6a-48c9-aef7-8480086aac11",
		Path:     "/",
		Domain:   "localhost:8000",
		Expires:  time.Now().Add(time.Hour * 24),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	ResponseWithJSON(w, http.StatusOK, nil)
}
