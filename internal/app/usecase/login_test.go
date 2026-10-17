package usecase_test

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"quizon/internal/app/usecase"
	"quizon/internal/pkg/cache"
)

func TestLogin(t *testing.T) {
	t.Parallel()

	const password = "secret"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}

	dbErr := errors.New("db is down")
	tests := []struct {
		name     string
		password string
		repoErr  error
		wantErr  error
	}{
		{name: "ok", password: password},
		{name: "wrong password", password: "nope", wantErr: usecase.ErrWrongPassword},
		// неизвестный логин снаружи неотличим от неверного пароля
		{name: "unknown login", password: password, repoErr: usecase.ErrNotFound, wantErr: usecase.ErrWrongPassword},
		{name: "db error", password: password, repoErr: dbErr, wantErr: dbErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := mockRepository(t)
			repo.EXPECT().GetPassword(gomock.Any(), "admin").Return(string(hash), tt.repoErr)
			u := usecase.NewUsecase(repo, cache.New[usecase.Session](), mockClock(testNow))

			session, err := u.Login(t.Context(), "admin", tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if session != (usecase.Session{}) {
					t.Errorf("failed login must not open a session: %+v", session)
				}
				return
			}

			// сессия на сутки по часам usecase, по её токену админа узнают
			if session.Token == "" || !session.ExpiresAt.Equal(testNow.Add(24*time.Hour)) {
				t.Errorf("session = %+v", session)
			}
			if !u.CheckSession(session.Token) {
				t.Error("new session must be valid")
			}
		})
	}
}
