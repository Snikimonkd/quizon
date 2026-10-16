package usecase

import (
	"testing"
	"time"

	"quizon/internal/generated/postgres/public/model"
)

func TestNumberRegistrations(t *testing.T) {
	t.Parallel()

	deletedAt := time.Now()
	got := numberRegistrations([]model.Registrations{
		{TeamName: "первые"},
		{TeamName: "удалённые", DeletedAt: &deletedAt},
		{TeamName: "третьи"},
	})

	// удалённая остаётся на своём месте, но номер не занимает
	want := []int64{1, 0, 2}
	for i, reg := range got {
		if reg.Number != want[i] || reg.Deleted() != (i == 1) {
			t.Errorf("%s: number = %d, deleted = %v", reg.TeamName, reg.Number, reg.Deleted())
		}
	}
}
