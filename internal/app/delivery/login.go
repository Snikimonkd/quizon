package delivery

import (
	"context"
	"net/http"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
)

type LoginUsecase interface {
	Login(ctx context.Context, name string, password string) (model.Cookie, error)
}

func (d delivery) Login(w http.ResponseWriter, r *http.Request) {
}
