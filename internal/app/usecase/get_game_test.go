package usecase

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"

	"quizon/internal/app/usecase/mocks"
	"quizon/internal/generated/postgres/public/model"
)

// mockRepository - мок репозитория без ожиданий: вызов, которого тест не задал, роняет тест
func mockRepository(t *testing.T) *mocks.MockRepository {
	t.Helper()

	return mocks.NewMockRepository(gomock.NewController(t))
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
		repo := mockRepository(t)
		repo.EXPECT().GetGame(gomock.Any(), game.ID).Return(game, nil)
		repo.EXPECT().RegistrationsAmount(gomock.Any(), game.ID).Return(tt.registered, nil)

		u := usecase{repository: repo}
		got, err := u.GetGame(t.Context(), game.ID)
		if err != nil || got.Games != game || got.Status != tt.want {
			t.Errorf("registered %d: GetGame = %+v, %v; want status %q", tt.registered, got, err, tt.want)
		}
	}

	// репозиторий оборачивает pgx.ErrNoRows, наружу уходит ErrNotFound. Регистрации у такой игры не считаются
	repo := mockRepository(t)
	notFound := fmt.Errorf("can't get game: %w", pgx.ErrNoRows)
	repo.EXPECT().GetGame(gomock.Any(), int64(8)).Return(model.Games{}, notFound)
	u := usecase{repository: repo}
	_, err := u.GetGame(t.Context(), 8)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("unknown game: err = %v, want ErrNotFound", err)
	}

	// без числа регистраций статус не узнать: ошибка, а не открытая регистрация
	dbErr := errors.New("db is down")
	repo = mockRepository(t)
	repo.EXPECT().GetGame(gomock.Any(), game.ID).Return(game, nil)
	repo.EXPECT().RegistrationsAmount(gomock.Any(), game.ID).Return(int64(0), dbErr)
	u = usecase{repository: repo}
	_, err = u.GetGame(t.Context(), game.ID)
	if !errors.Is(err, dbErr) {
		t.Errorf("count error: err = %v, want %v", err, dbErr)
	}
}
