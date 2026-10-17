package usecase

import (
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"quizon/internal/app/repository"
)

func TestListGames(t *testing.T) {
	t.Parallel()

	game := func(id int64, start, open time.Time, registered int64) repository.GameWithRegistrations {
		return repository.GameWithRegistrations{
			ID: id, MainAmount: 2, ReserveAmount: 1, StartTime: start, RegistrationOpenTime: open,
			RegistrationsAmount: registered,
		}
	}
	repo := mockRepository(t)
	repo.EXPECT().ListGames(gomock.Any(), int64(1), int64(100)).Return([]repository.GameWithRegistrations{
		game(1, testNow.Add(48*time.Hour), testNow.Add(time.Hour), 0),
		game(2, testNow.Add(48*time.Hour), testNow.Add(-time.Hour), 2),
		game(3, testNow.Add(-time.Hour), testNow.Add(-48*time.Hour), 3),
	}, nil)
	u := usecase{repository: repo, clock: mockClock(testNow)}

	got, err := u.ListGames(t.Context(), 1, 100)
	if err != nil {
		t.Fatalf("ListGames: %v", err)
	}
	// статус каждой игры — на момент запроса по часам usecase
	want := []RegistrationStatus{StatusWaiting, StatusReserve, StatusFinished}
	if len(got) != len(want) {
		t.Fatalf("games = %d, want %d", len(got), len(want))
	}
	for i, item := range got {
		if item.Status != want[i] {
			t.Errorf("game %d: status = %q, want %q", item.ID, item.Status, want[i])
		}
	}
}
