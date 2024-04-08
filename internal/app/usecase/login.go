package usecase

import (
	"context"
	"time"

	"github.com/Snikimonkd/quizon/internal/pkg/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type LoginRepository interface {
	Transactional(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
	GetPassword(ctx context.Context, tx pgx.Tx, name string) (string, error)
	Login(ctx context.Context, tx pgx.Tx, cookie model.Cookie) error
}

func (u usecase) Login(ctx context.Context, name string, password string) (model.Cookie, error) {
	cookie := model.Cookie{
		AdminName: name,
		Value:     uuid.New(),
		Expires:   time.Now().Add(time.Hour * 24 * 7),
	}
	err := u.loginRepository.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
		actualPassword, err := u.loginRepository.GetPassword(ctx, tx, name)
		if err != nil {
			return err
		}

		err = bcrypt.CompareHashAndPassword([]byte(actualPassword), []byte(password))
		if err != nil {
			return model.ErrWrongPassword
		}

		err = u.loginRepository.Login(ctx, tx, cookie)

		return nil
	})
	if err != nil {
		return model.Cookie{}, err
	}

	return cookie, nil
}
