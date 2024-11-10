package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	httpModel "quizon/internal/app/delivery/http/model"
	"quizon/internal/app/repository"
)

var ErrWrongPassword error = errors.New("wrong password")

type Cookie struct {
	Login     string
	Value     string
	ExpiresAt time.Time
}

type LoginRepository interface {
	GetPassword(ctx context.Context, login string) (string, error)
}

type CookieCache interface {
	Set(key string, val Cookie)
}

func (u usecase) Login(ctx context.Context, req httpModel.LoginRequest) (Cookie, error) {
	password, err := u.repository.GetPassword(ctx, req.Login)
	if errors.Is(err, repository.ErrNotFound) {
		return Cookie{}, ErrWrongPassword
	}
	if err != nil {
		return Cookie{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password))
	if err != nil {
		return Cookie{}, ErrWrongPassword
	}

	cookie := Cookie{
		Login:     req.Login,
		Value:     uuid.NewString(),
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}
	u.cookieCache.Set(cookie.Value, cookie)

	return cookie, nil
}
