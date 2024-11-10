package http

import "testing"

func Test_kek(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{
			name: "1. Successful test.",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			kek()
		})
	}
}
