package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	miniflux "miniflux.app/v2/client"
)

func makeEntry(title, content string, feedID int64, feedTitle, category string, published time.Time) *miniflux.Entry {
	return &miniflux.Entry{
		Title:   title,
		Content: content,
		URL:     "https://example.com/" + strings.ReplaceAll(strings.ToLower(title), " ", "-"),
		FeedID:  feedID,
		Date:    published,
		Feed: &miniflux.Feed{
			ID:    feedID,
			Title: feedTitle,
			Category: &miniflux.Category{
				Title: category,
			},
		},
	}
}

var testDate = time.Date(2026, 2, 24, 0, 0, 0, 0, time.UTC)

func TestStripHTML(t *testing.T) {
	got := stripHTML("<p>Hello &amp; <b>world</b></p>")
	want := "Hello & world"
	if got != want {
		t.Errorf("stripHTML = %q, want %q", got, want)
	}
}

func TestIsGossip(t *testing.T) {
	normal := makeEntry("Tech news today", "Some tech content", 1, "TechFeed", "news", testDate)
	if isGossip(normal) {
		t.Error("normal entry should not be gossip")
	}

	gossip := makeEntry("Ο γάμος του ζευγαριού", "Το ζευγάρι παντρεύτηκε", 1, "Feed", "news", testDate)
	if !isGossip(gossip) {
		t.Error("entry with multiple gossip keywords should be gossip")
	}

	lifoSingle := makeEntry("Ο γάμος της χρονιάς", "Some content", 96, "LiFO", "news", testDate)
	if !isGossip(lifoSingle) {
		t.Error("LiFO entry with single gossip keyword should be gossip")
	}
}

func TestExtractKeywords(t *testing.T) {
	kw := extractKeywords("The quick brown fox jumps over the lazy dog")
	if !kw["quick"] || !kw["brown"] || !kw["jumps"] || !kw["lazy"] {
		t.Errorf("unexpected keywords: %v", kw)
	}
	if kw["the"] || kw["over"] {
		t.Errorf("common/short words should be excluded: %v", kw)
	}
}

func TestAreSimilar(t *testing.T) {
	if !areSimilar(
		"Σεισμός 5.2 Ρίχτερ στην Κρήτη",
		"Ισχυρός σεισμός στην Κρήτη 5.2 Ρίχτερ",
	) {
		t.Error("similar titles should match")
	}

	if areSimilar("Σεισμός στην Κρήτη", "Αποτελέσματα εκλογών") {
		t.Error("unrelated titles should not match")
	}
}

func TestClusterEntries(t *testing.T) {
	entries := []*miniflux.Entry{
		makeEntry("Breaking: earthquake hits Crete", "content", 1, "Feed A", "news", testDate),
		makeEntry("Earthquake hits island of Crete", "content", 2, "Feed B", "news", testDate),
		makeEntry("Stock market rises today", "content", 3, "Feed C", "finance", testDate),
	}

	clusters := clusterEntries(entries)
	if len(clusters) != 1 {
		t.Fatalf("expected 1 cluster, got %d", len(clusters))
	}
	if len(clusters[0]) != 2 {
		t.Errorf("expected 2 entries in cluster, got %d", len(clusters[0]))
	}
}

func TestClusterEntriesSameFeed(t *testing.T) {
	entries := []*miniflux.Entry{
		makeEntry("Breaking: earthquake hits Crete", "content", 1, "Feed A", "news", testDate),
		makeEntry("Earthquake hits island of Crete", "content", 1, "Feed A", "news", testDate),
	}

	clusters := clusterEntries(entries)
	if len(clusters) != 0 {
		t.Errorf("entries from same feed should not cluster, got %d clusters", len(clusters))
	}
}

func TestProcessEntries(t *testing.T) {
	yesterday := testDate.AddDate(0, 0, -1)
	entries := []*miniflux.Entry{
		makeEntry("Breaking: earthquake hits Crete", "Long content about the earthquake", 1, "Feed A", "News", testDate),
		makeEntry("Earthquake hits island of Crete", "Another article about earthquake", 2, "Feed B", "News", testDate),
		makeEntry("Old news from yesterday", "content", 3, "Feed C", "News", yesterday),
		makeEntry("Celebrity wedding gossip γάμος ζευγάρι", "γάμος content ζευγάρι", 4, "Feed D", "News", testDate),
		makeEntry("Tech review of new phone", "content", 5, "Feed E", "Tech", testDate),
	}

	topics := processEntries(entries, testDate)
	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}
	if len(topics[0].sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(topics[0].sources))
	}
}

func TestFormatNews(t *testing.T) {
	topics := []hotTopic{
		{
			title:   "Breaking news about earthquake",
			summary: "A strong earthquake hit the region today",
			sources: []string{"Feed A", "Feed B", "Feed C"},
			url:     "https://example.com/earthquake",
		},
	}

	var buf bytes.Buffer
	formatNews(&buf, topics)
	out := buf.String()

	checks := []string{
		"=== News ===",
		"- Breaking news about earthquake",
		"https://example.com/earthquake",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}
}

func TestFormatNewsEmpty(t *testing.T) {
	var buf bytes.Buffer
	formatNews(&buf, nil)
	out := buf.String()

	if !strings.Contains(out, "no hot topics") {
		t.Errorf("expected 'no hot topics', got:\n%s", out)
	}
}

func TestBestSummary(t *testing.T) {
	entries := []*miniflux.Entry{
		{Content: "<p>Short.</p>"},
		{Content: "<p>First sentence here. Second sentence here. Third sentence here. Fourth sentence.</p>"},
	}
	s := bestSummary(entries)
	if !strings.Contains(s, "First sentence here") {
		t.Errorf("expected first sentence, got: %q", s)
	}
	if strings.Contains(s, "Fourth") {
		t.Errorf("should not include 4th sentence, got: %q", s)
	}
}
