package usecase_test

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"
)

// testGame - игра, которая проходит все правила: регистрация за двое суток до начала
func testGame() model.Games {
	return model.Games{
		Name: "Бауманская лига", Location: "МГТУ им. Н.Э.Баумана ауд. 345", MainAmount: 20, ReserveAmount: 10,
		StartTime: testNow.Add(72 * time.Hour), RegistrationOpenTime: testNow.Add(24 * time.Hour),
	}
}

func TestCreateGame(t *testing.T) {
	t.Parallel()

	// игра сохраняется со временем создания по часам usecase
	game := testGame()
	saved := game
	saved.CreatedAt = testNow
	repo := mockRepository(t)
	repo.EXPECT().CreateGame(gomock.Any(), saved).Return(int64(7), nil)
	u := usecase.NewUsecase(repo, nil, mockClock(testNow))
	id, err := u.CreateGame(t.Context(), game)
	if err != nil || id != 7 {
		t.Errorf("CreateGame = %d, %v; want 7", id, err)
	}
}

func TestCreateGameValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		edit  func(g *model.Games)
		field string
		rule  error
	}{
		{
			name: "no name", edit: func(g *model.Games) { g.Name = " " },
			field: usecase.FieldName, rule: usecase.ErrRequired,
		},
		{
			name: "no location", edit: func(g *model.Games) { g.Location = "" },
			field: usecase.FieldLocation, rule: usecase.ErrRequired,
		},
		{
			name: "no start", edit: func(g *model.Games) { g.StartTime = time.Time{} },
			field: usecase.FieldStartTime, rule: usecase.ErrRequired,
		},
		{
			name: "no opening", edit: func(g *model.Games) { g.RegistrationOpenTime = time.Time{} },
			field: usecase.FieldRegistrationOpenTime, rule: usecase.ErrRequired,
		},
		{
			name: "opens at start", edit: func(g *model.Games) { g.RegistrationOpenTime = g.StartTime },
			field: usecase.FieldRegistrationOpenTime, rule: usecase.ErrOpensAfterStart,
		},
		{
			name: "opens after start", edit: func(g *model.Games) { g.RegistrationOpenTime = g.StartTime.Add(time.Hour) },
			field: usecase.FieldRegistrationOpenTime, rule: usecase.ErrOpensAfterStart,
		},
		{
			name: "no main places", edit: func(g *model.Games) { g.MainAmount = 0 },
			field: usecase.FieldMainAmount, rule: usecase.ErrTooSmall,
		},
		{
			name: "negative reserve", edit: func(g *model.Games) { g.ReserveAmount = -1 },
			field: usecase.FieldReserveAmount, rule: usecase.ErrTooSmall,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			game := testGame()
			tt.edit(&game)
			// игру с нарушением не сохраняем: CreateGame мок не ждёт
			u := usecase.NewUsecase(mockRepository(t), nil, mockClock(testNow))

			_, err := u.CreateGame(t.Context(), game)
			// поле с нарушением видно через errors.As, само правило — через errors.Is
			var invalid usecase.ValidationError
			if !errors.As(err, &invalid) || len(invalid) != 1 || !errors.Is(invalid[tt.field], tt.rule) ||
				!errors.Is(err, tt.rule) {
				t.Errorf("err = %v, want only %s: %v", err, tt.field, tt.rule)
			}
		})
	}
}
