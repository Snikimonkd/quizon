package usecase_test

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/utils"
)

func TestListRegistrations(t *testing.T) {
	t.Parallel()

	// вторая команда удалена: остаётся на своём месте в списке, но номер не занимает
	deletedAt := testNow.Add(time.Hour)
	regs := []model.Registrations{
		{ID: 10, TeamName: "первые", CreatedAt: testNow},
		{ID: 20, TeamName: "удалённые", CreatedAt: testNow.Add(time.Minute), DeletedAt: &deletedAt},
		{ID: 30, TeamName: "третьи", CreatedAt: testNow.Add(2 * time.Minute)},
	}
	repo := mockRepository(t)
	repo.EXPECT().ListRegistrations(gomock.Any(), int64(7)).Return(regs, nil)
	u := usecase.NewUsecase(repo, nil, mockClock(testNow))

	got, err := u.ListRegistrations(t.Context(), 7)
	if err != nil {
		t.Fatalf("ListRegistrations: %v", err)
	}
	want := []int64{1, 0, 2}
	if len(got) != len(want) {
		t.Fatalf("registrations = %d, want %d", len(got), len(want))
	}
	for i, reg := range got {
		// порядок подачи сохраняется
		if reg.ID != regs[i].ID || reg.Number != want[i] {
			t.Errorf("%s: id = %d, number = %d; want id %d, number %d",
				reg.TeamName, reg.ID, reg.Number, regs[i].ID, want[i])
		}
		// время подачи переводится в московское, момент тот же
		if reg.CreatedAt.Location() != utils.LocMsk || !reg.CreatedAt.Equal(regs[i].CreatedAt) {
			t.Errorf("%s: created at %v, want %v in Moscow", reg.TeamName, reg.CreatedAt, regs[i].CreatedAt)
		}
	}

	// ошибку репозитория отдаёт как есть
	dbErr := errors.New("db is down")
	repo = mockRepository(t)
	repo.EXPECT().ListRegistrations(gomock.Any(), int64(7)).Return(nil, dbErr)
	u = usecase.NewUsecase(repo, nil, mockClock(testNow))
	_, err = u.ListRegistrations(t.Context(), 7)
	if !errors.Is(err, dbErr) {
		t.Errorf("repository error: err = %v, want %v", err, dbErr)
	}
}
