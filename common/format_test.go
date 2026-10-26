package common

import (
	"math"
	"testing"
)

func TestFormatCountInt64(t *testing.T) {
	cases := []struct {
		count int64
		want  string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{999999, "1000.0K"},
		{1000000, "1.0M"},
		{2500000, "2.5M"},
		{1000000000, "1.0B"},
		{1000000000000, "1.0T"},
	}
	for _, c := range cases {
		if got := FormatCountInt64(c.count); got != c.want {
			t.Errorf("FormatCountInt64(%d) = %q, want %q", c.count, got, c.want)
		}
	}
}

func TestFormatCountFloat64(t *testing.T) {
	cases := []struct {
		count float64
		want  string
	}{
		{0, "0.0"},
		{999.5, "999.5"},
		{1000, "1.0K"},
		{1000000, "1.0M"},
	}
	for _, c := range cases {
		if got := FormatCountFloat64(c.count); got != c.want {
			t.Errorf("FormatCountFloat64(%v) = %q, want %q", c.count, got, c.want)
		}
	}
}

func TestFormatTime(t *testing.T) {
	cases := []struct {
		t    int64
		want string
	}{
		{0, "0ms"},
		{999, "999ms"},
		{1000, "1.0s"},
		{1000000, "1.0min"},
		{1000000000, "1.0hour"},
		{1000000000000, "1.0day"},
	}
	for _, c := range cases {
		if got := FormatTime(c.t); got != c.want {
			t.Errorf("FormatTime(%d) = %q, want %q", c.t, got, c.want)
		}
	}
}

func TestFormatByte(t *testing.T) {
	cases := []struct {
		b    int64
		want string
	}{
		{0, "0B"},
		{999, "999B"},
		{1000, "1.0KB"},
		{1500, "1.5KB"},
		{1000000, "1.0MB"},
		{1000000000, "1.0GB"},
		{1000000000000, "1.0TB"},
	}
	for _, c := range cases {
		if got := FormatByte(c.b); got != c.want {
			t.Errorf("FormatByte(%d) = %q, want %q", c.b, got, c.want)
		}
	}
	// 极大值不应产生 +Inf 之类的非法输出
	if got := FormatByte(math.MaxInt64); got == "" {
		t.Error("FormatByte(MaxInt64) returned empty string")
	}
}
