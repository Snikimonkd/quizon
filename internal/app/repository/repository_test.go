package repository

import (
	"context"
	"testing"

	"quizon/internal/generated/postgres/public/model"
	"quizon/internal/pkg/testsupport"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
				login: "123",
			},
			want: func(ctx context.Context, t *testing.T, db *pgxpool.Pool, login string) string {
				pass := uuid.NewString()
				hashPass, err := bcrypt.GenerateFromPassword([]byte(pass), 12)
				if err != nil {
					t.Errorf("can't generate password: %v", err)
				}
				testsupport.InsertIntoAdmins(ctx, t, db, model.Admins{
					Login:    login,
					Password: string(hashPass),
				})
				return string(hashPass)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
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
			if !cmp.Equal(got, want) {
				t.Errorf("repository.GetPassword(), diff: %v", cmp.Diff(got, want))
			}
		})
	}
}
