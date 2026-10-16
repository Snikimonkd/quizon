package usecase

import "testing"

func TestRegistrationStatus(t *testing.T) {
	t.Parallel()

	// основной состав 2, резерв 1
	tests := []struct {
		registered int64
		want       RegistrationStatus
	}{
		{registered: 0, want: StatusOk},
		{registered: 1, want: StatusOk},
		{registered: 2, want: StatusReserve},
		{registered: 3, want: StatusClosed},
		{registered: 4, want: StatusClosed},
	}
	for _, tt := range tests {
		got := registrationStatus(tt.registered, 2, 1)
		if got != tt.want {
			t.Errorf("registrationStatus(%d, 2, 1) = %q, want %q", tt.registered, got, tt.want)
		}
	}

	if got := registrationStatus(0, 0, 0); got != StatusClosed {
		t.Errorf("game without places: %q, want %q", got, StatusClosed)
	}
}
