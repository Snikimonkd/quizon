package usecase

import (
	"errors"

	"github.com/benbjohnson/clock"
)

// Repository - интерфейс инкапсулирующий в себе все репозитории
type Repository interface {
	//	LoginRepository
	RegisterRepository
	RegistrationsRepository
	RegisterAvailableRepository
	CreateGameRepository
}

type usecase struct {
	//	loginRepository             LoginRepository
	repository Repository
	clock      clock.Clock
}

// NewUsecase - конструктор для usecase
func NewUsecase(repository Repository) usecase {
	return usecase{
		//	loginRepository:             repositories,
		repository: repository,
		clock:      clock.New(),
	}
}

// ErrNotFound - ошибка "не найдено"
var ErrNotFound = errors.New("not found error")
