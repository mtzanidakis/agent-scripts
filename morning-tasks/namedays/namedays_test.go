package namedays

import (
	"strings"
	"testing"
	"time"
)

func TestOrthodoxEaster(t *testing.T) {
	// Known Orthodox Easter dates (Gregorian)
	cases := []struct {
		year int
		want string
	}{
		{2024, "2024-05-05"},
		{2025, "2025-04-20"},
		{2026, "2026-04-12"},
		{2027, "2027-05-02"},
		{2028, "2028-04-16"},
		{2030, "2030-04-28"},
	}
	for _, tc := range cases {
		got := OrthodoxEaster(tc.year).Format("2006-01-02")
		if got != tc.want {
			t.Errorf("OrthodoxEaster(%d) = %s, want %s", tc.year, got, tc.want)
		}
	}
}

func TestForDateRecurring(t *testing.T) {
	// Nov 8 is Αγγελος nameday (08/11)
	date := time.Date(2026, time.November, 8, 0, 0, 0, 0, time.UTC)
	names, err := ForDate(date)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("expected names for November 8")
	}
	found := false
	for _, n := range names {
		if n == "Αγγελος" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Αγγελος in names, got %v", names)
	}
}

func TestForDateEasterRelative(t *testing.T) {
	// Easter 2026 is April 12. Easter day itself (offset 0) should have Αναστάσιος etc.
	date := time.Date(2026, time.April, 12, 0, 0, 0, 0, time.UTC)
	names, err := ForDate(date)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, n := range names {
		if n == "Αναστάσιος" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected Αναστάσιος on Easter, got %v", names)
	}
}

func TestFormatPrefix(t *testing.T) {
	// April 24 has Αχιλλέας
	date := time.Date(2026, time.April, 24, 0, 0, 0, 0, time.UTC)
	out, err := Format(date)
	if err != nil {
		t.Fatal(err)
	}
	want := "=== Namedays ===\nγιορτάζουν σήμερα: "
	if len(out) < len(want) || out[:len(want)] != want {
		t.Errorf("expected prefix %q, got %q", want, out)
	}
}

func TestFormat(t *testing.T) {
	date := time.Date(2026, time.November, 8, 0, 0, 0, 0, time.UTC)
	out, err := Format(date)
	if err != nil {
		t.Fatal(err)
	}
	if out == "" {
		t.Fatal("expected non-empty format output")
	}
	if !strings.Contains(out, "=== Namedays ===") {
		t.Errorf("expected title, got: %q", out)
	}
	if len(out) < len("=== Namedays ===\nγιορτάζουν σήμερα: ") {
		t.Errorf("unexpected format: %q", out)
	}
}
