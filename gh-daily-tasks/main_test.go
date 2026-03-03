package main

import (
	"testing"
	"time"
)

func TestBuildOutputSkipsInactiveRepos(t *testing.T) {
	reports := []RepoReport{
		{FullName: "user/active", Stars: 10, Forks: 2, OpenPRs: []PR{{Number: 1, Title: "fix"}}},
		{FullName: "user/inactive", Stars: 5, Forks: 1},
	}
	deltas := map[string][2]int{
		"user/active":   {0, 0},
		"user/inactive": {0, 0},
	}

	out := buildOutput(reports, deltas)

	if len(out.Repos) != 1 {
		t.Fatalf("expected 1 active repo, got %d", len(out.Repos))
	}
	if out.Repos[0].Name != "user/active" {
		t.Errorf("expected user/active, got %s", out.Repos[0].Name)
	}
	if out.TotalRepos != 2 {
		t.Errorf("TotalRepos should count all repos (2), got %d", out.TotalRepos)
	}
}

func TestBuildOutputIncludesRepoWithDeltaOnly(t *testing.T) {
	reports := []RepoReport{
		{FullName: "user/starred", Stars: 10, Forks: 2},
	}
	deltas := map[string][2]int{
		"user/starred": {3, 0}, // gained 3 stars, no other activity
	}

	out := buildOutput(reports, deltas)

	if len(out.Repos) != 1 {
		t.Fatalf("expected 1 repo (has star delta), got %d", len(out.Repos))
	}
	if out.Repos[0].Stars.Delta != 3 {
		t.Errorf("star delta: got %d, want 3", out.Repos[0].Stars.Delta)
	}
}

func TestBuildOutputSummary(t *testing.T) {
	now := time.Now()
	reports := []RepoReport{
		{
			FullName: "user/repo1",
			Stars:    10, Forks: 2,
			OpenPRs:    []PR{{Number: 1}, {Number: 2}},
			OpenIssues: []Issue{{Number: 3, CreatedAt: now}},
			DependabotAlerts: []DependabotAlert{
				{Number: 1, SecurityAdvisory: SecurityAdvisory{Severity: "critical", Summary: "bad"}},
				{Number: 2, SecurityAdvisory: SecurityAdvisory{Severity: "high", Summary: "also bad"}},
			},
		},
		{
			FullName: "user/repo2",
			Stars:    5, Forks: 1,
			OpenPRs: []PR{{Number: 10}},
			DependabotAlerts: []DependabotAlert{
				{Number: 3, SecurityAdvisory: SecurityAdvisory{Severity: "critical", Summary: "worst"}},
			},
		},
	}
	deltas := map[string][2]int{
		"user/repo1": {2, 1},
		"user/repo2": {-1, 0},
	}

	out := buildOutput(reports, deltas)

	s := out.Summary
	if s.TotalOpenPRs != 3 {
		t.Errorf("TotalOpenPRs: got %d, want 3", s.TotalOpenPRs)
	}
	if s.TotalOpenIssues != 1 {
		t.Errorf("TotalOpenIssues: got %d, want 1", s.TotalOpenIssues)
	}
	if s.TotalAlerts != 3 {
		t.Errorf("TotalAlerts: got %d, want 3", s.TotalAlerts)
	}
	if s.AlertsBySeverity["critical"] != 2 {
		t.Errorf("critical alerts: got %d, want 2", s.AlertsBySeverity["critical"])
	}
	if s.AlertsBySeverity["high"] != 1 {
		t.Errorf("high alerts: got %d, want 1", s.AlertsBySeverity["high"])
	}
	if s.NetStarsDelta != 1 {
		t.Errorf("NetStarsDelta: got %d, want 1", s.NetStarsDelta)
	}
	if s.NetForksDelta != 1 {
		t.Errorf("NetForksDelta: got %d, want 1", s.NetForksDelta)
	}
}

func TestBuildOutputPRFields(t *testing.T) {
	created := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	reports := []RepoReport{
		{
			FullName: "user/repo",
			OpenPRs: []PR{
				{Number: 42, Title: "Fix login", User: User{Login: "alice"}, HTMLURL: "https://github.com/user/repo/pull/42", CreatedAt: created, Draft: true},
			},
		},
	}
	deltas := map[string][2]int{"user/repo": {0, 0}}

	out := buildOutput(reports, deltas)

	if len(out.Repos) != 1 || len(out.Repos[0].OpenPRs) != 1 {
		t.Fatal("expected 1 repo with 1 PR")
	}
	pr := out.Repos[0].OpenPRs[0]
	if pr.Number != 42 {
		t.Errorf("Number: got %d", pr.Number)
	}
	if pr.Author != "alice" {
		t.Errorf("Author: got %q", pr.Author)
	}
	if !pr.Draft {
		t.Error("Draft should be true")
	}
	if pr.URL != "https://github.com/user/repo/pull/42" {
		t.Errorf("URL: got %q", pr.URL)
	}
}

func TestBuildOutputIssueLabels(t *testing.T) {
	now := time.Now()
	reports := []RepoReport{
		{
			FullName: "user/repo",
			OpenIssues: []Issue{
				{Number: 5, Title: "Bug", User: User{Login: "bob"}, CreatedAt: now, Labels: []Label{{Name: "bug"}, {Name: "urgent"}}},
			},
		},
	}
	deltas := map[string][2]int{"user/repo": {0, 0}}

	out := buildOutput(reports, deltas)

	if len(out.Repos) != 1 || len(out.Repos[0].OpenIssues) != 1 {
		t.Fatal("expected 1 repo with 1 issue")
	}
	iss := out.Repos[0].OpenIssues[0]
	if len(iss.Labels) != 2 || iss.Labels[0] != "bug" || iss.Labels[1] != "urgent" {
		t.Errorf("Labels: got %v, want [bug urgent]", iss.Labels)
	}
}

func TestBuildOutputEmptyReports(t *testing.T) {
	out := buildOutput(nil, nil)

	if out.TotalRepos != 0 {
		t.Errorf("TotalRepos: got %d", out.TotalRepos)
	}
	if out.Repos != nil {
		t.Errorf("Repos should be nil, got %v", out.Repos)
	}
	if out.Summary.AlertsBySeverity == nil {
		t.Error("AlertsBySeverity should be initialized")
	}
}
