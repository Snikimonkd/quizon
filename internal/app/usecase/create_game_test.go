package usecase_test

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"
)

func TestCreateGame(t *testing.T) {
	t.Parallel()

	game := model.Games{
		Name: "Бауманская лига", MainAmount: 20, ReserveAmount: 10,
		StartTime: testNow.Add(72 * time.Hour), RegistrationOpenTime: testNow.Add(24 * time.Hour),
	}

	// игра сохраняется со временем создания по часам usecase
	saved := game
	saved.CreatedAt = testNow
	repo := mockRepository(t)
	repo.EXPECT().CreateGame(gomock.Any(), saved).Return(int64(7), nil)
	u := usecase.NewUsecase(repo, nil, mockClock(testNow))
	id, err := u.CreateGame(t.Context(), game)
	if err != nil || id != 7 {
		t.Errorf("CreateGame = %d, %v; want 7", id, err)
	}

	// регистрация открывается не раньше начала игры: такую игру не сохраняем, CreateGame мок не ждёт
	for _, open := range []time.Time{game.StartTime, game.StartTime.Add(time.Hour)} {
		bad := game
		bad.RegistrationOpenTime = open
		u := usecase.NewUsecase(mockRepository(t), nil, mockClock(testNow))
		_, err := u.CreateGame(t.Context(), bad)
		if !errors.Is(err, usecase.ErrOpensAfterStart) {
			t.Errorf("opens at %v: err = %v, want ErrOpensAfterStart", open, err)
		}
	}
}
