package usecase_test

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"quizon/internal/app/usecase"
)

func TestDeleteRegistration(t *testing.T) {
	t.Parallel()

	// удаление — под той же блокировкой игры, что и регистрация, со временем по часам usecase
	repo := mockRepository(t)
	expectTx(repo)
	gomock.InOrder(
		repo.EXPECT().LockGame(gomock.Any(), gomock.Any(), int64(7)).Return(nil),
		repo.EXPECT().DeleteRegistration(gomock.Any(), gomock.Any(), int64(7), int64(20), testNow).Return(nil),
	)
	u := usecase.NewUsecase(repo, nil, mockClock(testNow))
	err := u.DeleteRegistration(t.Context(), 7, 20)
	if err != nil {
		t.Errorf("DeleteRegistration: %v", err)
	}

	// команду уже удалили: ErrNotFound репозитория уходит наружу как есть
	repo = mockRepository(t)
	expectTx(repo)
	repo.EXPECT().LockGame(gomock.Any(), gomock.Any(), int64(7)).Return(nil)
	repo.EXPECT().DeleteRegistration(gomock.Any(), gomock.Any(), int64(7), int64(20), testNow).
		Return(usecase.ErrNotFound)
	u = usecase.NewUsecase(repo, nil, mockClock(testNow))
	err = u.DeleteRegistration(t.Context(), 7, 20)
	if !errors.Is(err, usecase.ErrNotFound) {
		t.Errorf("already deleted: err = %v, want ErrNotFound", err)
	}
}
