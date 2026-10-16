package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/pkg/testsupport"

	"github.com/google/uuid"
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
				login: uuid.NewString(),
			},
			want: func(ctx context.Context, t *testing.T, db *pgxpool.Pool, login string) string {
				t.Helper()
				pass := uuid.NewString()
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
		Location:             "ауд. 345",
		Name:                 uuid.NewString(),
		MainAmount:           1,
		ReserveAmount:        1,
		RegistrationOpenTime: now,
	})

	ids := make([]int64, 0, 3)
	for i := range 3 {
		ids = append(ids, testsupport.InsertRegistration(ctx, t, db, model.Registrations{
			GameID:        gameID,
			CreatedAt:     now.Add(time.Duration(i) * time.Second),
			TeamName:      uuid.NewString(),
			CaptainName:   "капитан",
			Phone:         "89990000000",
			Telegram:      "@team",
			PlayersAmount: 5,
		}))
	}

	err := deleteRegistration(ctx, r, gameID, ids[1])
	if err != nil {
		t.Fatalf("can't delete registration: %v", err)
	}

	return gameID, ids
}

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
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("second delete: err = %v, want ErrNotFound", err)
	}
	err = deleteRegistration(ctx, r, gameID+1, ids[0])
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("delete from another game: err = %v, want ErrNotFound", err)
	}

	regs, err := r.ListRegistrations(ctx, gameID)
	if err != nil {
		t.Fatalf("can't list registrations: %v", err)
	}
	if len(regs) != 2 || regs[0].ID != ids[0] || regs[1].ID != ids[2] {
		t.Errorf("ListRegistrations must skip deleted: %+v", regs)
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

	var cnt int64
	err = r.Transactional(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var txErr error
		_, _, _, cnt, txErr = r.GetRegistrationsAmount(ctx, tx, gameID)
		return txErr
	})
	if err != nil || cnt != 2 {
		t.Errorf("GetRegistrationsAmount = %d, %v; want 2", cnt, err)
	}

	games, err := r.ListGames(ctx, 1, 100)
	if err != nil {
		t.Fatalf("can't list games: %v", err)
	}
	game, ok := lo.Find(games, func(g GameWithRegistrations) bool { return g.ID == gameID })
	if !ok || game.RegistrationsAmount != 2 {
		t.Errorf("ListGames: found = %v, registrations = %d; want 2", ok, game.RegistrationsAmount)
	}
}
