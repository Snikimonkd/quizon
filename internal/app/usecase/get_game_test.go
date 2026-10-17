package usecase_test

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"
)

func TestGetGame(t *testing.T) {
	t.Parallel()

	// основной состав 2, резерв 1, регистрация открыта, игра через три дня:
	// статус по оставшимся в игре командам, как в списке игр
	game := model.Games{
		ID: 7, Name: "Бауманская лига", MainAmount: 2, ReserveAmount: 1,
		StartTime: testNow.Add(72 * time.Hour), RegistrationOpenTime: testNow.Add(-time.Hour),
	}
	tests := []struct {
		registered int64
		want       usecase.RegistrationStatus
	}{
		{registered: 1, want: usecase.StatusOk},
		{registered: 2, want: usecase.StatusReserve},
		{registered: 3, want: usecase.StatusClosed},
	}
	for _, tt := range tests {
		repo := mockRepository(t)
		repo.EXPECT().GetGame(gomock.Any(), game.ID).Return(game, nil)
		repo.EXPECT().RegistrationsAmount(gomock.Any(), game.ID).Return(tt.registered, nil)

		u := usecase.NewUsecase(repo, nil, mockClock(testNow))
		got, err := u.GetGame(t.Context(), game.ID)
		if err != nil || got.Games != game || got.Status != tt.want {
			t.Errorf("registered %d: GetGame = %+v, %v; want status %q", tt.registered, got, err, tt.want)
		}
	}

	// время берётся с часов usecase: после начала игры регистрации нет, даже если места есть
	repo := mockRepository(t)
	repo.EXPECT().GetGame(gomock.Any(), game.ID).Return(game, nil)
	repo.EXPECT().RegistrationsAmount(gomock.Any(), game.ID).Return(int64(0), nil)
	u := usecase.NewUsecase(repo, nil, mockClock(game.StartTime))
	got, err := u.GetGame(t.Context(), game.ID)
	if err != nil || got.Status != usecase.StatusFinished {
		t.Errorf("after start: status = %q, %v; want %q", got.Status, err, usecase.StatusFinished)
	}

	// игры нет: ErrNotFound репозитория уходит наружу как есть, регистрации у такой игры не считаются
	repo = mockRepository(t)
	repo.EXPECT().GetGame(gomock.Any(), int64(8)).Return(model.Games{}, usecase.ErrNotFound)
	u = usecase.NewUsecase(repo, nil, mockClock(testNow))
	_, err = u.GetGame(t.Context(), 8)
	if !errors.Is(err, usecase.ErrNotFound) {
		t.Errorf("unknown game: err = %v, want ErrNotFound", err)
	}

	// без числа регистраций статус не узнать: ошибка, а не открытая регистрация
	dbErr := errors.New("db is down")
	repo = mockRepository(t)
	repo.EXPECT().GetGame(gomock.Any(), game.ID).Return(game, nil)
	repo.EXPECT().RegistrationsAmount(gomock.Any(), game.ID).Return(int64(0), dbErr)
	u = usecase.NewUsecase(repo, nil, mockClock(testNow))
	_, err = u.GetGame(t.Context(), game.ID)
	if !errors.Is(err, dbErr) {
		t.Errorf("count error: err = %v, want %v", err, dbErr)
	}
}
