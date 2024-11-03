package usecase

import (
	"errors"

	"github.com/benbjohnson/clock"
)

// Repositories - интерфейс инкапсулирующий в себе все репозитории
type Repositories interface {
	//	LoginRepository
	RegisterRepository
	RegistrationsRepository
	RegisterAvailableRepository
}

type usecase struct {
	//	loginRepository             LoginRepository
	registerRepository          RegisterRepository
	registrationsRepository     RegistrationsRepository
	registerAvailableRepository RegisterAvailableRepository
	clock                       clock.Clock
}

// NewUsecase - конструктор для usecase
func NewUsecase(repositories Repositories) usecase {
	return usecase{
		//	loginRepository:             repositories,
		registerRepository:          repositories,
		registrationsRepository:     repositories,
		registerAvailableRepository: repositories,

		clock: clock.New(),
	}
}

// ErrNotFound - ошибка "не найдено"
var ErrNotFound = errors.New("not found error")
