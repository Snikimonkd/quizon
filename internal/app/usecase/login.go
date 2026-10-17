package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrWrongPassword error = errors.New("wrong password")

// sessionTTL - сколько действует сессия админа после входа
const sessionTTL = time.Hour * 24

type LoginRepository interface {
	// GetPassword - bcrypt-хеш пароля админа. ErrNotFound, если такого логина нет
	GetPassword(ctx context.Context, login string) (string, error)
}

// Login - проверяет логин и пароль админа и открывает сессию.
// ErrWrongPassword, если такого логина нет или пароль неверный
func (u usecase) Login(ctx context.Context, login, password string) (Session, error) {
	hash, err := u.repository.GetPassword(ctx, login)
	if errors.Is(err, ErrNotFound) {
		return Session{}, ErrWrongPassword
	}
	if err != nil {
		return Session{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return Session{}, ErrWrongPassword
	}

	session := Session{Token: uuid.NewString(), ExpiresAt: u.clock.Now().Add(sessionTTL)}
	u.sessions.Set(session.Token, session)

	return session, nil
}
