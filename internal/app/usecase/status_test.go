package usecase

import (
	"testing"
	"time"

	"quizon/internal/generated/postgres/public/model"
)

// testNow - точка отсчёта времени в тестах статусов. Тесты на моках живут во внешнем пакете usecase_test,
// а этим нужны неэкспортируемые функции, поэтому у них своё время
var testNow = time.Date(2026, time.October, 16, 9, 0, 0, 0, time.UTC)

func TestRegistrationStatus(t *testing.T) {
	t.Parallel()

	// основной состав 2, резерв 1
	tests := []struct {
		registered int64
		want       RegistrationStatus
	}{
		{registered: 0, want: StatusOk},
		{registered: 1, want: StatusOk},
		{registered: 2, want: StatusReserve},
		{registered: 3, want: StatusClosed},
		{registered: 4, want: StatusClosed},
	}
	for _, tt := range tests {
		got := registrationStatus(tt.registered, 2, 1)
		if got != tt.want {
			t.Errorf("registrationStatus(%d, 2, 1) = %q, want %q", tt.registered, got, tt.want)
		}
	}

	if got := registrationStatus(0, 0, 0); got != StatusClosed {
		t.Errorf("game without places: %q, want %q", got, StatusClosed)
	}
}

func TestGameStatus(t *testing.T) {
	t.Parallel()

	// основной состав 2, резерв 1; регистрация открывается за сутки до игры
	game := model.Games{
		MainAmount: 2, ReserveAmount: 1,
		StartTime: testNow.Add(24 * time.Hour), RegistrationOpenTime: testNow,
	}
	tests := []struct {
		name       string
		now        time.Time
		registered int64
		want       RegistrationStatus
	}{
		{name: "before opening", now: testNow.Add(-time.Second), want: StatusWaiting},
		{name: "opening moment", now: testNow, want: StatusOk},
		{name: "reserve", now: testNow, registered: 2, want: StatusReserve},
		{name: "no places", now: testNow, registered: 3, want: StatusClosed},
		// время важнее мест: у начавшейся игры места обычно разобраны, но она уже прошла
		{name: "game started, full", now: game.StartTime, registered: 3, want: StatusFinished},
		{name: "game started, free", now: game.StartTime.Add(time.Hour), want: StatusFinished},
	}
	for _, tt := range tests {
		if got := gameStatus(game, tt.registered, tt.now); got != tt.want {
			t.Errorf("%s: gameStatus = %q, want %q", tt.name, got, tt.want)
		}
	}
}
