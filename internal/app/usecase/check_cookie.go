package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
)

type CheckCookieRepository interface {
	CheckCookie(ctx context.Context, cookie string) (time.Time, error)
}

func (u usecase) CheckCookie(ctx context.Context, cookie string) (bool, error) {
	expires, err := u.checkAuthRepository.CheckCookie(ctx, cookie)
	if errors.Is(err, model.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return time.Now().Before(expires), nil
}
