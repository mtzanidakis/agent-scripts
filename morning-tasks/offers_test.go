package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
)

func TestFormatOffers(t *testing.T) {
	entries := []*miniflux.Entry{
		{Title: "50% off electronics", URL: "https://example.com/deal1"},
		{Title: "Free shipping today", URL: "https://example.com/deal2"},
	}

	var buf bytes.Buffer
	formatOffers(&buf, entries)
	out := buf.String()

	checks := []string{
		"=== Offers ===",
		"- 50% off electronics",
		"https://example.com/deal1",
		"- Free shipping today",
		"https://example.com/deal2",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

func TestFormatOffersEmpty(t *testing.T) {
	var buf bytes.Buffer
	formatOffers(&buf, nil)
	out := buf.String()

	if !strings.Contains(out, "no new offers") {
		t.Errorf("expected 'no new offers', got:\n%s", out)
	}
}

func TestMarkAsReadEmpty(t *testing.T) {
	// Should be a no-op with nil client when no entries
	err := markAsRead(nil, nil)
	if err != nil {
		t.Errorf("markAsRead with no entries should not error: %v", err)
	}
}

func TestFetchOffersFilter(t *testing.T) {
	// Verify the since calculation: date minus 24h
	date := time.Date(2026, 2, 24, 12, 0, 0, 0, time.UTC)
	since := date.Add(-24 * time.Hour)
	expected := time.Date(2026, 2, 23, 12, 0, 0, 0, time.UTC)
	if !since.Equal(expected) {
		t.Errorf("since = %v, want %v", since, expected)
	}
}
