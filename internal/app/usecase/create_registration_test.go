package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"

	"quizon/internal/app/usecase"
	"quizon/internal/app/usecase/mock"
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

func TestCreateRegistrationErrors(t *testing.T) {
	t.Parallel()

	// места есть: до ошибок дело дошло бы до записи
	game := usecase.GameWithRegistrations{
		ID: 7, MainAmount: 2, ReserveAmount: 1,
		StartTime: testNow.Add(72 * time.Hour), RegistrationOpenTime: testNow.Add(-time.Hour),
	}
	dbErr := errors.New("db is down")
	lock := func(repo *mock.MockRepository, err error) {
		repo.EXPECT().LockGame(gomock.Any(), gomock.Any(), int64(7)).Return(err)
	}
	read := func(repo *mock.MockRepository, game usecase.GameWithRegistrations, err error) {
		repo.EXPECT().GetGameWithRegistrations(gomock.Any(), gomock.Any(), int64(7)).Return(game, err)
	}

	// после ошибки следующие шаги не выполняются: их мок не ждёт
	tests := []struct {
		name   string
		expect func(repo *mock.MockRepository)
		want   error
	}{
		{
			// без блокировки места не считаем
			name: "lock failed",
			expect: func(repo *mock.MockRepository) {
				expectTx(repo)
				lock(repo, dbErr)
			},
			want: dbErr,
		},
		{
			name: "game not found",
			expect: func(repo *mock.MockRepository) {
				expectTx(repo)
				lock(repo, nil)
				read(repo, usecase.GameWithRegistrations{}, usecase.ErrNotFound)
			},
			want: usecase.ErrNotFound,
		},
		{
			name: "insert failed",
			expect: func(repo *mock.MockRepository) {
				expectTx(repo)
				lock(repo, nil)
				read(repo, game, nil)
				repo.EXPECT().CreateRegistration(gomock.Any(), gomock.Any(), gomock.Any()).Return(dbErr)
			},
			want: dbErr,
		},
		{
			// запись прошла, а коммит нет: команды в базе нет, и успех показывать нельзя,
			// хотя статус внутри транзакции уже посчитан
			name: "commit failed",
			expect: func(repo *mock.MockRepository) {
				repo.EXPECT().Transactional(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
						err := fn(ctx, nil)
						if err != nil {
							return err
						}

						return dbErr
					},
				)
				lock(repo, nil)
				read(repo, game, nil)
				repo.EXPECT().CreateRegistration(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
			want: dbErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := mockRepository(t)
			tt.expect(repo)
			u := usecase.NewUsecase(repo, nil, mockClock(testNow))

			got, err := u.CreateRegistration(t.Context(), model.Registrations{GameID: 7, TeamName: "КвизON"})
			if !errors.Is(err, tt.want) || got != "" {
				t.Errorf("CreateRegistration = %q, %v; want no status, %v", got, err, tt.want)
			}
		})
	}
}
