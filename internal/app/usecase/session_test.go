package usecase_test

import (
	"testing"
	"time"

	"quizon/internal/app/usecase"
	"quizon/internal/pkg/cache"
)

func TestCheckSession(t *testing.T) {
	t.Parallel()

	sessions := cache.New[usecase.Session]()
	sessions.Set("live", usecase.Session{Token: "live", ExpiresAt: testNow.Add(time.Hour)})
	sessions.Set("expired", usecase.Session{Token: "expired", ExpiresAt: testNow.Add(-time.Second)})
	u := usecase.NewUsecase(nil, sessions, mockClock(testNow))

	if !u.CheckSession("live") {
		t.Error("live session must be valid")
	}
	if u.CheckSession("unknown") {
		t.Error("unknown token must be rejected")
	}

	// срок сверяется с часами usecase, истёкшая сессия удаляется
	if u.CheckSession("expired") {
		t.Error("expired session must be rejected")
	}
	if _, ok := sessions.Get("expired"); ok {
		t.Error("expired session must be deleted")
	}
}

func TestLogout(t *testing.T) {
	t.Parallel()

	sessions := cache.New[usecase.Session]()
	sessions.Set("token", usecase.Session{Token: "token", ExpiresAt: testNow.Add(time.Hour)})
	u := usecase.NewUsecase(nil, sessions, mockClock(testNow))

	u.Logout("token")
	if u.CheckSession("token") {
		t.Error("session must end after logout")
	}
}
