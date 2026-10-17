package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"

	"quizon/internal/app/usecase/mock"
)

// testNow - время на часах usecase в тестах
var testNow = time.Date(2026, time.October, 16, 9, 0, 0, 0, time.UTC)

// mockClock - часы, которые стоят на at
func mockClock(at time.Time) *clock.Mock {
	c := clock.NewMock()
	c.Set(at)

	return c
}

// mockRepository - мок репозитория без ожиданий: вызов, которого тест не задал, роняет тест
func mockRepository(t *testing.T) *mock.MockRepository {
	t.Helper()

	return mock.NewMockRepository(gomock.NewController(t))
}

// expectTx - Transactional мока сразу выполняет fn: настоящая транзакция тестам usecase не нужна
func expectTx(repo *mock.MockRepository) {
	repo.EXPECT().Transactional(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context, pgx.Tx) error) error {
			return fn(ctx, nil)
		},
	)
}
