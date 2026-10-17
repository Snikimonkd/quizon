package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"quizon/internal/generated/postgres/public/model"

	"github.com/jackc/pgx/v5"
)

// gameRepository - одна игра и число её регистраций; остальные методы репозитория GetGame не нужны
type gameRepository struct {
	Repository

	game       model.Games
	registered int64
	countErr   error
}

func (r gameRepository) GetGame(_ context.Context, gameID int64) (model.Games, error) {
	if gameID != r.game.ID {
		return model.Games{}, fmt.Errorf("can't get game: %w", pgx.ErrNoRows)
	}

	return r.game, nil
}

func (r gameRepository) RegistrationsAmount(_ context.Context, gameID int64) (int64, error) {
	// как в базе: у другой игры регистраций нет
	if gameID != r.game.ID {
		return 0, nil
	}

	return r.registered, r.countErr
}

func TestGetGame(t *testing.T) {
	t.Parallel()

	// основной состав 2, резерв 1: статус по оставшимся в игре командам, как в списке игр
	game := model.Games{ID: 7, Name: "Бауманская лига", MainAmount: 2, ReserveAmount: 1}
	tests := []struct {
		registered int64
		want       RegistrationStatus
	}{
		{registered: 1, want: StatusOk},
		{registered: 2, want: StatusReserve},
		{registered: 3, want: StatusClosed},
	}
	for _, tt := range tests {
		u := usecase{repository: gameRepository{game: game, registered: tt.registered}}
		got, err := u.GetGame(t.Context(), game.ID)
		if err != nil || got.Games != game || got.Status != tt.want {
			t.Errorf("registered %d: GetGame = %+v, %v; want status %q", tt.registered, got, err, tt.want)
		}
	}

	u := usecase{repository: gameRepository{game: game}}
	_, err := u.GetGame(t.Context(), game.ID+1)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown game: err = %v, want ErrNotFound", err)
	}

	// без числа регистраций статус не узнать: ошибка, а не открытая регистрация
	dbErr := errors.New("db is down")
	u = usecase{repository: gameRepository{game: game, countErr: dbErr}}
	_, err = u.GetGame(t.Context(), game.ID)
	if !errors.Is(err, dbErr) {
		t.Errorf("count error: err = %v, want %v", err, dbErr)
	}
}
