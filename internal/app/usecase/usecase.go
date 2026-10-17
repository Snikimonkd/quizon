package usecase

import (
	"context"
	"errors"

	"github.com/benbjohnson/clock"
	"github.com/jackc/pgx/v5"
)

//go:generate go tool mockgen -typed -destination=mock/repository_mock.go -package=mock . Repository

// Repository - интерфейс инкапсулирующий в себе все репозитории
type Repository interface {
	TxRepository
	ListGamesRepository
	GetGameRepository
	ListRegistrationsRepository

	RegisterRepository
	DeleteRegistrationRepository
	CreateGameRepository
	LoginRepository
}

// TxRepository - транзакции и блокировка игры. Что делать внутри транзакции, решает usecase:
// репозиторий только выполняет запросы
type TxRepository interface {
	// Transactional - выполняет fn в одной транзакции: коммит, если fn вернула nil, иначе откат
	Transactional(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
	// LockGame - блокирует игру до конца транзакции: регистрации и удаления в одной игре идут по очереди
	LockGame(ctx context.Context, tx pgx.Tx, gameID int64) error
}

type usecase struct {
	repository Repository
	clock      clock.Clock

	sessions SessionStore
}

// NewUsecase - конструктор для usecase. clk - часы: в main настоящие, в тестах стоят на месте
func NewUsecase(repository Repository, sessions SessionStore, clk clock.Clock) usecase {
	return usecase{
		repository: repository,
		clock:      clk,
		sessions:   sessions,
	}
}

// ErrNotFound - нет такой игры, регистрации или админа. Репозиторий отдаёт её вместо ошибок базы
var ErrNotFound = errors.New("not found error")
