package pricing

import "testing"

func TestMoneyString(t *testing.T) {
	tests := []struct {
		in   Money
		want string
	}{
		{Rub(150), "150.00 ₽"},
		{Money(5), "0.05 ₽"},
		{Money(-1050), "-10.50 ₽"},
		{Money(0), "0.00 ₽"},
	}
	for _, tt := range tests {
		if got := tt.in.String(); got != tt.want {
			t.Errorf("Money(%d).String() = %q, want %q", int64(tt.in), got, tt.want)
		}
	}
}

func TestMoneyRoundToRuble(t *testing.T) {
	tests := []struct {
		in, want Money
	}{
		{14949, Rub(149)},
		{14950, Rub(150)},
		{Rub(150), Rub(150)},
	}
	for _, tt := range tests {
		if got := tt.in.RoundToRuble(); got != tt.want {
			t.Errorf("Money(%d).RoundToRuble() = %v, want %v", int64(tt.in), got, tt.want)
		}
	}
}
