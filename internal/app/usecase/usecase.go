package usecase

import (
	"errors"

	"github.com/benbjohnson/clock"
)

// Repository - интерфейс инкапсулирующий в себе все репозитории
type Repository interface {
	RegisterRepository
	RegistrationsRepository
	RegisterAvailableRepository
	CreateGameRepository
	ListGamesRepository
	LoginRepository
}

type usecase struct {
	//	loginRepository             LoginRepository
	repository Repository
	clock      clock.Clock

	cookieCache CookieCache
}

// NewUsecase - конструктор для usecase
func NewUsecase(repository Repository, cookieCache CookieCache) usecase {
	return usecase{
		repository:  repository,
		clock:       clock.New(),
		cookieCache: cookieCache,
	}
}

// ErrNotFound - ошибка "не найдено"
var ErrNotFound = errors.New("not found error")
