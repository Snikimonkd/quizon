package usecase

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"quizon/internal/app/delivery/api"
	"quizon/internal/app/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrWrongPassword error = errors.New("wrong password")

const AuthorizationTokenName string = `authorization-token`

type LoginRepository interface {
	GetPassword(ctx context.Context, login string) (string, error)
}

type CookieCache interface {
	Set(key string, val *http.Cookie)
}

func (u usecase) Login(ctx context.Context, req api.PostLoginRequestObject) (*http.Cookie, error) {
	password, err := u.repository.GetPassword(ctx, req.Body.Login)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrWrongPassword
	}
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Body.Password))
	if err != nil {
		return nil, ErrWrongPassword
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		return nil, fmt.Errorf("can't find domain in env variable")
	}

	cookie := &http.Cookie{
		Name:    AuthorizationTokenName,
		Value:   uuid.NewString(),
		Expires: time.Now().Add(time.Hour * 24),

		Path:   "/",
		Domain: domain,
		// https
		Secure: true,
		// only visible to browser and not to js
		HttpOnly: true,
		// send from local host to domain
		SameSite: http.SameSiteLaxMode,
	}

	u.cookieCache.Set(cookie.Value, cookie)

	return cookie, nil
}
