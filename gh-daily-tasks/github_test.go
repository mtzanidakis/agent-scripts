package main

import (
	"bytes"
	"io"
	"testing"
)

func TestFilterRepos(t *testing.T) {
	repos := []Repo{
		{FullName: "user/good", Owner: User{Login: "user"}},
		{FullName: "user/archived", Owner: User{Login: "user"}, Archived: true},
		{FullName: "user/forked", Owner: User{Login: "user"}, Fork: true},
		{FullName: "org/not-mine", Owner: User{Login: "org"}},
		{FullName: "user/also-good", Owner: User{Login: "user"}, StargazersCount: 5},
	}

	var filtered []Repo
	for _, r := range repos {
		if r.Archived || r.Fork || r.Owner.Login != "user" {
			continue
		}
		filtered = append(filtered, r)
	}

	if len(filtered) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(filtered))
	}
	if filtered[0].FullName != "user/good" {
		t.Errorf("first repo: got %s", filtered[0].FullName)
	}
	if filtered[1].FullName != "user/also-good" {
		t.Errorf("second repo: got %s", filtered[1].FullName)
	}
}

func TestFilterIssuesExcludesPRs(t *testing.T) {
	prLink := &struct{}{}
	issues := []Issue{
		{Number: 1, Title: "Real issue"},
		{Number: 2, Title: "Actually a PR", PullRequestLinks: prLink},
		{Number: 3, Title: "Another issue"},
	}

	var filtered []Issue
	for _, iss := range issues {
		if iss.PullRequestLinks == nil {
			filtered = append(filtered, iss)
		}
	}

	if len(filtered) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(filtered))
	}
	if filtered[0].Number != 1 || filtered[1].Number != 3 {
		t.Errorf("got issues %d and %d, want 1 and 3", filtered[0].Number, filtered[1].Number)
	}
}

func TestLogWarning(t *testing.T) {
	var buf bytes.Buffer
	old := warningWriter
	warningWriter = &buf
	defer func() { warningWriter = old }()

	logWarning("something %s: %d", "broke", 42)

	got := buf.String()
	want := "warning: something broke: 42\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestLogWarningDiscard(t *testing.T) {
	old := warningWriter
	warningWriter = io.Discard
	defer func() { warningWriter = old }()

	// Should not panic
	logWarning("test %s", "message")
}
