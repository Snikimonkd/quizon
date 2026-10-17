package usecase

import (
	"net/http"
	"testing"
	"time"

	"quizon/internal/pkg/cache"
)

// sessionCookie - сессия админа, которая истекает в expires
func sessionCookie(expires time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     AuthorizationTokenName,
		Expires:  expires,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func TestCheckSession(t *testing.T) {
	t.Parallel()

	sessions := cache.New[*http.Cookie]()
	sessions.Set("live", sessionCookie(testNow.Add(time.Hour)))
	sessions.Set("expired", sessionCookie(testNow.Add(-time.Second)))
	u := usecase{cookieCache: sessions, clock: mockClock(testNow)}

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

	sessions := cache.New[*http.Cookie]()
	sessions.Set("token", sessionCookie(testNow.Add(time.Hour)))
	u := usecase{cookieCache: sessions, clock: mockClock(testNow)}

	u.Logout("token")
	if u.CheckSession("token") {
		t.Error("session must end after logout")
	}
}
