package usecase_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"
)

func TestListRegistrations(t *testing.T) {
	t.Parallel()

	// в основе одно место; вторая команда удалена: остаётся на своём месте, но номер и место не занимает
	deletedAt := testNow.Add(time.Hour)
	regs := []model.Registrations{
		{ID: 10, TeamName: "первые", CreatedAt: testNow},
		{ID: 20, TeamName: "удалённые", CreatedAt: testNow.Add(time.Minute), DeletedAt: &deletedAt},
		{ID: 30, TeamName: "третьи", CreatedAt: testNow.Add(2 * time.Minute)},
		{ID: 40, TeamName: "четвёртые", CreatedAt: testNow.Add(3 * time.Minute)},
	}
	repo := mockRepository(t)
	repo.EXPECT().GetGame(gomock.Any(), int64(7)).Return(model.Games{ID: 7, MainAmount: 1, ReserveAmount: 2}, nil)
	repo.EXPECT().ListRegistrations(gomock.Any(), int64(7)).Return(regs, nil)
	u := usecase.NewUsecase(repo, nil, mockClock(testNow))

	got, err := u.ListRegistrations(t.Context(), 7)
	if err != nil {
		t.Fatalf("ListRegistrations: %v", err)
	}
	// порядок подачи и сами регистрации как есть; первая оставшаяся за основой — первая в резерве
	want := []usecase.Registration{
		{Registrations: regs[0], Number: 1},
		{Registrations: regs[1], Deleted: true},
		{Registrations: regs[2], Number: 2, Reserve: true, FirstReserve: true},
		{Registrations: regs[3], Number: 3, Reserve: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ListRegistrations =\n%+v\nwant\n%+v", got, want)
	}
}

func TestListRegistrationsErrors(t *testing.T) {
	t.Parallel()

	// игры нет: регистрации не запрашиваются
	repo := mockRepository(t)
	repo.EXPECT().GetGame(gomock.Any(), int64(8)).Return(model.Games{}, usecase.ErrNotFound)
	u := usecase.NewUsecase(repo, nil, mockClock(testNow))
	_, err := u.ListRegistrations(t.Context(), 8)
	if !errors.Is(err, usecase.ErrNotFound) {
		t.Errorf("unknown game: err = %v, want ErrNotFound", err)
	}

	// ошибку репозитория отдаёт как есть
	dbErr := errors.New("db is down")
	repo = mockRepository(t)
	repo.EXPECT().GetGame(gomock.Any(), int64(7)).Return(model.Games{ID: 7, MainAmount: 1}, nil)
	repo.EXPECT().ListRegistrations(gomock.Any(), int64(7)).Return(nil, dbErr)
	u = usecase.NewUsecase(repo, nil, mockClock(testNow))
	_, err = u.ListRegistrations(t.Context(), 7)
	if !errors.Is(err, dbErr) {
		t.Errorf("repository error: err = %v, want %v", err, dbErr)
	}
}
