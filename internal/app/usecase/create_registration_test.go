package usecase_test

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"
)

func TestCreateRegistration(t *testing.T) {
	t.Parallel()

	// основной состав 2, резерв 1, регистрация открыта, игра через три дня
	open := model.Games{
		ID: 7, MainAmount: 2, ReserveAmount: 1,
		StartTime: testNow.Add(72 * time.Hour), RegistrationOpenTime: testNow.Add(-time.Hour),
	}
	waiting := open
	waiting.RegistrationOpenTime = testNow.Add(time.Hour)
	started := open
	started.StartTime = testNow.Add(-time.Hour)

	tests := []struct {
		name       string
		game       model.Games
		registered int64
		want       usecase.RegistrationStatus
		wantErr    error
		wantSave   bool
	}{
		{name: "main", game: open, registered: 1, want: usecase.StatusOk, wantSave: true},
		{name: "reserve", game: open, registered: 2, want: usecase.StatusReserve, wantSave: true},
		{name: "no places", game: open, registered: 3, want: usecase.StatusClosed},
		{name: "not opened yet", game: waiting, wantErr: usecase.ErrNotOpenedYet},
		// места есть, но игра уже идёт: регистрации нет, как и в списке игр
		{name: "game started", game: started, want: usecase.StatusFinished},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			in := model.Registrations{GameID: 7, TeamName: "КвизON"}
			saved := in
			saved.CreatedAt = testNow

			// места считаются под блокировкой игры, регистрация сохраняется, только если они есть
			repo := mockRepository(t)
			expectTx(repo)
			calls := []any{
				repo.EXPECT().LockGame(gomock.Any(), gomock.Any(), int64(7)).Return(nil),
				repo.EXPECT().GetGameWithRegistrations(gomock.Any(), gomock.Any(), int64(7)).
					Return(usecase.GameWithRegistrations{Games: tt.game, RegistrationsAmount: tt.registered}, nil),
			}
			if tt.wantSave {
				calls = append(calls, repo.EXPECT().CreateRegistration(gomock.Any(), gomock.Any(), saved).Return(nil))
			}
			gomock.InOrder(calls...)
			u := usecase.NewUsecase(repo, nil, mockClock(testNow))

			got, err := u.CreateRegistration(t.Context(), in)
			if !errors.Is(err, tt.wantErr) || got != tt.want {
				t.Errorf("CreateRegistration = %q, %v; want %q, %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}
