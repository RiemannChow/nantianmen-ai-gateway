package costopt

import "testing"

func TestRound2(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{1.234, 1.23},
		{1.235, 1.24},
		{0, 0},
		{1234.5678, 1234.57},
	}
	for _, c := range cases {
		if got := round2(c.in); got != c.want {
			t.Errorf("round2(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestBudgetLevelLabel(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"ok", "正常"},
		{"warning", "已达预警线"},
		{"exceeded", "已超预算"},
		{"", "正常"},
		{"unknown", "正常"},
	}
	for _, c := range cases {
		if got := budgetLevelLabel(c.in); got != c.want {
			t.Errorf("budgetLevelLabel(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
