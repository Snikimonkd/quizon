package usecase

import "testing"

func TestValidPhone(t *testing.T) {
	t.Parallel()

	// считаются только цифры: от 10 до 15
	tests := []struct {
		phone string
		want  bool
	}{
		{phone: "999 888-77-6", want: false},
		{phone: "999 888-77-66", want: true},
		{phone: "8 (999) 888-77-66", want: true},
		{phone: "+44 20 7946 0958 123", want: true},
		{phone: "+44 20 7946 0958 1234", want: false},
		{phone: "", want: false},
	}
	for _, tt := range tests {
		if got := validPhone(tt.phone); got != tt.want {
			t.Errorf("validPhone(%q) = %v, want %v", tt.phone, got, tt.want)
		}
	}
}

func TestValidationErrorText(t *testing.T) {
	t.Parallel()

	// поля в тексте по алфавиту: строки журнала одинаковые при одинаковых нарушениях
	err := ValidationError{FieldTeamName: ErrRequired, FieldPhone: ErrBadPhone}
	if got, want := err.Error(), "invalid Phone: bad phone, TeamName: required"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
