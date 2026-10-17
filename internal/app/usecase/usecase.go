package usecase

import (
	"errors"

	"github.com/benbjohnson/clock"
)

//go:generate go tool mockgen -typed -destination=mocks/repository_mock.go -package=mocks . Repository

// Repository - интерфейс инкапсулирующий в себе все репозитории
type Repository interface {
	ListGamesRepository
	GetGameRepository
	ListRegistrationsRepository

	RegisterRepository
	DeleteRegistrationRepository
	CreateGameRepository
	LoginRepository
}

type usecase struct {
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
