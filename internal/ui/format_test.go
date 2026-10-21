package ui

import (
	"testing"
	"unicode/utf8"
)

func TestHumanSize(t *testing.T) {
	cases := map[int64]string{
		0:               "0 B",
		1023:            "1023 B",
		1024:            "1.0 KiB",
		1536:            "1.5 KiB",
		5 * 1024 * 1024: "5.0 MiB",
		3 << 30:         "3.0 GiB",
		2 << 40:         "2.0 TiB",
	}
	for in, want := range cases {
		if got := HumanSize(in); got != want {
			t.Errorf("HumanSize(%d) = %q, esperado %q", in, got, want)
		}
	}
}

func TestBar(t *testing.T) {
	cases := []struct {
		frac float64
		want string
	}{
		{0, "    "},
		{1, "████"},
		{0.5, "██  "},
		{0.125, "▌   "},
		{2, "████"},
		{-1, "    "},
	}
	for _, c := range cases {
		got := Bar(c.frac, 4)
		if got != c.want {
			t.Errorf("Bar(%v, 4) = %q, esperado %q", c.frac, got, c.want)
		}
		if n := utf8.RuneCountInString(got); n != 4 {
			t.Errorf("Bar(%v, 4) tem %d colunas", c.frac, n)
		}
	}
}
