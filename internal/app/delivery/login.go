package delivery

import (
	"context"
	"errors"
	"net/http"

	"quizon/internal/app/delivery/api"
	"quizon/internal/app/usecase"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type LoginUsecase interface {
	Login(ctx context.Context, req api.PostLoginRequestObject) (*http.Cookie, error)
}

type loginResponse struct {
	cookie *http.Cookie
}

func (l loginResponse) VisitPostLoginResponse(w http.ResponseWriter) error {
	http.SetCookie(w, l.cookie)
	return nil
}

func (d *delivery) PostLogin(
	ctx context.Context,
	req api.PostLoginRequestObject,
) (api.PostLoginResponseObject, error) {
	err := validateLoginRequest(req.Body)
	if err != nil {
		return api.PostLogin400JSONResponse{
			Error: err.Error(),
		}, nil
	}

	cookie, err := d.usecase.Login(ctx, req)
	if errors.Is(err, usecase.ErrWrongPassword) {
		return api.PostLogin400JSONResponse{
			Error: err.Error(),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	resp := loginResponse{
		cookie: cookie,
	}

	return resp, nil
}

func validateLoginRequest(req *api.PostLoginJSONRequestBody) error {
	return validation.ValidateStruct(
		req,
		validation.Field(&req.Login, validation.Required),
		validation.Field(&req.Password, validation.Required),
	)
}
