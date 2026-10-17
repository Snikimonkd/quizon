package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"quizon/internal/app/usecase"
	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/pkg/testsupport"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"
	"golang.org/x/crypto/bcrypt"
)

func Test_repository_GetPassword(t *testing.T) {
	t.Parallel()

	type fields struct {
		db func(ctx context.Context, t *testing.T) *pgxpool.Pool
	}
	type args struct {
		ctx   context.Context
		login string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(ctx context.Context, t *testing.T, db *pgxpool.Pool, login string) string
		wantErr bool
	}{
		{
			name: "1. Successful test",
			fields: fields{
				db: testsupport.ConnectToTestPostgres,
			},
			args: args{
				ctx:   context.Background(),
				login: gofakeit.Username(),
			},
			want: func(ctx context.Context, t *testing.T, db *pgxpool.Pool, login string) string {
				t.Helper()
				pass := gofakeit.Password(true, true, true, true, false, 16)
				hashPass, err := bcrypt.GenerateFromPassword([]byte(pass), 12)
				if err != nil {
					t.Errorf("can't generate password: %v", err)
				}
				testsupport.InsertIntoAdmins(ctx, t, db, model.Admins{
					Login:    login,
					Password: string(hashPass),
				})
				return pass
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := repository{
				db: tt.fields.db(tt.args.ctx, t),
			}

			want := tt.want(tt.args.ctx, t, r.db, tt.args.login)

			got, err := r.GetPassword(
				tt.args.ctx,
				tt.args.login,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("repository.GetPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			err = bcrypt.CompareHashAndPassword([]byte(got), []byte(want))
			if err != nil {
				t.Errorf("want != got, want = %v, got = %v, err = %v", want, got, err)
			}
		})
	}
}

// gameWithDeletedTeam - игра с тремя командами, вторая из которых удалена.
// Возвращает id игры и id регистраций в порядке подачи.
func gameWithDeletedTeam(ctx context.Context, t *testing.T, db *pgxpool.Pool) (int64, []int64) {
	t.Helper()

	r := NewRepository(db)
	now := time.Now()
	gameID := testsupport.InsertGame(ctx, t, db, model.Games{
		CreatedAt: now,
		// далеко в будущем, чтобы игра точно попала на первую страницу ListGames
		StartTime:            now.AddDate(100, 0, 0),
		Location:             "МГТУ им. Н.Э.Баумана ауд. 345",
		Name:                 gofakeit.Company(),
		MainAmount:           1,
		ReserveAmount:        1,
		RegistrationOpenTime: now,
	})

	ids := make([]int64, 0, 3)
	for i := range 3 {
		reg := testsupport.FakeRegistration(gameID)
		reg.CreatedAt = now.Add(time.Duration(i) * time.Second)
		ids = append(ids, testsupport.InsertRegistration(ctx, t, db, reg))
	}

	err := deleteRegistration(ctx, r, gameID, ids[1])
	if err != nil {
		t.Fatalf("can't delete registration: %v", err)
	}

	return gameID, ids
}

// deleteRegistration - удаляет регистрацию в своей транзакции, как это делает usecase
func deleteRegistration(ctx context.Context, r repository, gameID, registrationID int64) error {
	return r.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return r.DeleteRegistration(ctx, tx, gameID, registrationID, time.Now())
	})
}

func Test_repository_DeleteRegistration(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	db := testsupport.ConnectToTestPostgres(ctx, t)
	r := NewRepository(db)
	gameID, ids := gameWithDeletedTeam(ctx, t, db)

	// повторно и из чужой игры удалить нельзя
	err := deleteRegistration(ctx, r, gameID, ids[1])
	if !errors.Is(err, usecase.ErrNotFound) {
		t.Errorf("second delete: err = %v, want ErrNotFound", err)
	}
	err = deleteRegistration(ctx, r, gameID+1, ids[0])
	if !errors.Is(err, usecase.ErrNotFound) {
		t.Errorf("delete from another game: err = %v, want ErrNotFound", err)
	}

	regs, err := r.ListRegistrations(ctx, gameID)
	if err != nil {
		t.Fatalf("can't list registrations: %v", err)
	}
	// удалённая остаётся на своём месте в списке, но с deleted_at
	if len(regs) != 3 || regs[1].ID != ids[1] || regs[1].DeletedAt == nil ||
		regs[0].DeletedAt != nil || regs[2].DeletedAt != nil {
		t.Errorf("ListRegistrations must keep deleted with deleted_at: %+v", regs)
	}
}

func Test_repository_CountsSkipDeleted(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	db := testsupport.ConnectToTestPostgres(ctx, t)
	r := NewRepository(db)
	gameID, _ := gameWithDeletedTeam(ctx, t, db)

	amount, err := r.RegistrationsAmount(ctx, gameID)
	if err != nil || amount != 2 {
		t.Errorf("RegistrationsAmount = %d, %v; want 2", amount, err)
	}

	// в транзакции регистрации игра приходит вместе с числом оставшихся команд
	var locked usecase.GameWithRegistrations
	err = r.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var txErr error
		locked, txErr = r.GetGameWithRegistrations(ctx, tx, gameID)
		return txErr
	})
	if err != nil || locked.ID != gameID || locked.RegistrationsAmount != 2 {
		t.Errorf("GetGameWithRegistrations = game %d with %d registrations, %v; want game %d with 2",
			locked.ID, locked.RegistrationsAmount, err, gameID)
	}

	games, err := r.ListGames(ctx, 1, 100)
	if err != nil {
		t.Fatalf("can't list games: %v", err)
	}
	game, ok := lo.Find(games, func(g usecase.GameWithRegistrations) bool { return g.ID == gameID })
	if !ok || game.RegistrationsAmount != 2 {
		t.Errorf("ListGames: found = %v, registrations = %d; want 2", ok, game.RegistrationsAmount)
	}
}

func Test_repository_Transactional(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	db := testsupport.ConnectToTestPostgres(ctx, t)
	r := NewRepository(db)
	gameID, ids := gameWithDeletedTeam(ctx, t, db)

	// register - сохраняет новую команду в транзакции, которая заканчивается ошибкой fail
	register := func(fail error) (model.Registrations, error) {
		reg := testsupport.FakeRegistration(gameID)
		reg.CreatedAt = time.Now().Add(time.Hour)

		return reg, r.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
			err := r.CreateRegistration(ctx, tx, reg)
			if err != nil {
				return err
			}

			return fail
		})
	}

	// ошибка откатывает запись и возвращается как есть
	errStop := errors.New("stop")
	_, err := register(errStop)
	if !errors.Is(err, errStop) {
		t.Errorf("failed tx: err = %v, want %v", err, errStop)
	}

	// без ошибки запись остаётся и встаёт в конец списка
	reg, err := register(nil)
	if err != nil {
		t.Fatalf("can't register: %v", err)
	}

	regs, err := r.ListRegistrations(ctx, gameID)
	if err != nil {
		t.Fatalf("can't list registrations: %v", err)
	}
	if len(regs) != len(ids)+1 || regs[len(regs)-1].TeamName != reg.TeamName {
		t.Errorf("only the committed registration must be saved: %+v", regs)
	}
}

// Test_repository_NotFound - вместо ошибок базы репозиторий отдаёт usecase.ErrNotFound
func Test_repository_NotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	r := NewRepository(testsupport.ConnectToTestPostgres(ctx, t))

	_, err := r.GetGame(ctx, -1)
	if !errors.Is(err, usecase.ErrNotFound) {
		t.Errorf("GetGame: err = %v, want ErrNotFound", err)
	}
	_, err = r.GetPassword(ctx, "nobody-"+gofakeit.UUID())
	if !errors.Is(err, usecase.ErrNotFound) {
		t.Errorf("GetPassword: err = %v, want ErrNotFound", err)
	}
	err = r.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, txErr := r.GetGameWithRegistrations(ctx, tx, -1)
		return txErr
	})
	if !errors.Is(err, usecase.ErrNotFound) {
		t.Errorf("GetGameWithRegistrations: err = %v, want ErrNotFound", err)
	}
}
