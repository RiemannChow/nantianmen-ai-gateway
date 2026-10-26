package multitenant

import "testing"

func TestBudgetLevel(t *testing.T) {
	cases := []struct {
		ratio float64
		want  string
	}{
		{0, "ok"},
		{0.5, "ok"},
		{0.79, "ok"},
		{0.8, "warning"},
		{0.95, "warning"},
		{1.0, "exceeded"},
		{1.5, "exceeded"},
	}
	for _, c := range cases {
		if got := budgetLevel(c.ratio, 0.8); got != c.want {
			t.Errorf("budgetLevel(%v, 0.8) = %q, want %q", c.ratio, got, c.want)
		}
	}
}

func TestRound2(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{1.234, 1.23},
		{1.235, 1.24},
		{0, 0},
		{2.5, 2.5},
		{99.999, 100.0},
	}
	for _, c := range cases {
		if got := round2(c.in); got != c.want {
			t.Errorf("round2(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}
