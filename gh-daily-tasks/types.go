package main

import "time"

// GitHub API response types

type User struct {
	Login string `json:"login"`
}

type Repo struct {
	FullName        string `json:"full_name"`
	Owner           User   `json:"owner"`
	Archived        bool   `json:"archived"`
	Fork            bool   `json:"fork"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
}

type PR struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	User      User      `json:"user"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
	Draft     bool      `json:"draft"`
}

type Issue struct {
	Number           int       `json:"number"`
	Title            string    `json:"title"`
	User             User      `json:"user"`
	HTMLURL          string    `json:"html_url"`
	CreatedAt        time.Time `json:"created_at"`
	Labels           []Label   `json:"labels"`
	PullRequestLinks *struct{} `json:"pull_request,omitempty"` // non-nil means it's a PR
}

type Label struct {
	Name string `json:"name"`
}

type DependabotAlert struct {
	Number                int                    `json:"number"`
	State                 string                 `json:"state"`
	SecurityAdvisory      SecurityAdvisory       `json:"security_advisory"`
	SecurityVulnerability *SecurityVulnerability `json:"security_vulnerability,omitempty"`
	HTMLURL               string                 `json:"html_url"`
}

type SecurityAdvisory struct {
	Summary  string `json:"summary"`
	Severity string `json:"severity"`
}

type SecurityVulnerability struct {
	Severity string `json:"severity"`
}

// Per-repo collected data

type RepoReport struct {
	FullName         string
	Stars            int
	Forks            int
	OpenPRs          []PR
	OpenIssues       []Issue
	DependabotAlerts []DependabotAlert
}

// Star/Fork delta

type Delta struct {
	Current int `json:"current"`
	Delta   int `json:"delta"`
}

// JSON output types

type OutputReport struct {
	Date       string        `json:"date"`
	TotalRepos int           `json:"total_repos"`
	Repos      []OutputRepo  `json:"repos"`
	Summary    OutputSummary `json:"summary"`
}

type OutputRepo struct {
	Name             string        `json:"name"`
	OpenPRs          []OutputPR    `json:"open_prs,omitempty"`
	OpenIssues       []OutputIssue `json:"open_issues,omitempty"`
	DependabotAlerts []OutputAlert `json:"dependabot_alerts,omitempty"`
	Stars            Delta         `json:"stars"`
	Forks            Delta         `json:"forks"`
}

type OutputPR struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
	Draft     bool   `json:"draft"`
}

type OutputIssue struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	Author    string   `json:"author"`
	URL       string   `json:"url"`
	CreatedAt string   `json:"created_at"`
	Labels    []string `json:"labels,omitempty"`
}

type OutputAlert struct {
	Number   int    `json:"number"`
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	URL      string `json:"url"`
}

type OutputSummary struct {
	TotalOpenPRs     int            `json:"total_open_prs"`
	TotalOpenIssues  int            `json:"total_open_issues"`
	TotalAlerts      int            `json:"total_alerts"`
	AlertsBySeverity map[string]int `json:"alerts_by_severity"`
	NetStarsDelta    int            `json:"net_stars_delta"`
	NetForksDelta    int            `json:"net_forks_delta"`
}
