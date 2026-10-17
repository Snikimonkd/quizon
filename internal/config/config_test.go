package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	const (
		domain string = "quiz-on.ru"
		dsn    string = "postgres://postgres@localhost:5432/postgres"
	)

	tests := []struct {
		name string
		// debug - задана ли переменная DEBUG
		debug   bool
		domain  string
		dsn     string
		want    Config
		wantErr string
	}{
		{
			name:   "prod",
			domain: domain,
			dsn:    dsn,
			want:   Config{Domain: domain, PgDSN: dsn},
		},
		{
			name:   "debug",
			debug:  true,
			domain: domain,
			dsn:    dsn,
			want:   Config{Debug: true, Domain: domain, PgDSN: dsn},
		},
		{name: "no domain", dsn: dsn, wantErr: envDomain},
		{name: "no dsn", domain: domain, wantErr: envPgDSN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envDomain, tt.domain)
			t.Setenv(envPgDSN, tt.dsn)
			// значение DEBUG не важно, даже пустое включает отладку, поэтому без неё переменную удаляем.
			// t.Setenv вернёт прежнее значение после теста и в этом случае
			t.Setenv(envDebug, "")
			if !tt.debug {
				err := os.Unsetenv(envDebug)
				if err != nil {
					t.Fatal(err)
				}
			}

			got, err := Load()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want error about %s", err, tt.wantErr)
				}

				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("config = %+v, want %+v", got, tt.want)
			}
		})
	}
}
