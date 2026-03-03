package main

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/cli/go-gh/v2/pkg/api"
)

func fetchAuthenticatedUser(client *api.RESTClient) (string, error) {
	var user User
	err := client.Get("user", &user)
	if err != nil {
		return "", fmt.Errorf("fetching authenticated user (run `gh auth login`): %w", err)
	}
	return user.Login, nil
}

// paginateGet fetches all pages from a paginated GitHub API endpoint.
// The path must already include per_page=100.
func paginateGet[T any](client *api.RESTClient, basePath string) ([]T, error) {
	var all []T
	page := 1
	for {
		path := fmt.Sprintf("%s&page=%d", basePath, page)
		var items []T
		if err := client.Get(path, &items); err != nil {
			return nil, checkRateLimit(err, path)
		}
		all = append(all, items...)
		if len(items) < 100 {
			break
		}
		page++
	}
	return all, nil
}

func checkRateLimit(err error, context string) error {
	var httpErr *api.HTTPError
	if errors.As(err, &httpErr) {
		if httpErr.StatusCode == 403 || httpErr.StatusCode == 429 {
			resetTime := httpErr.Headers.Get("X-RateLimit-Reset")
			if resetTime != "" {
				return fmt.Errorf("rate limited on %s; resets at unix timestamp %s", context, resetTime)
			}
		}
	}
	return fmt.Errorf("%s: %w", context, err)
}

func fetchAllRepos(client *api.RESTClient, username string) ([]Repo, error) {
	allRepos, err := paginateGet[Repo](client, "user/repos?per_page=100&type=owner")
	if err != nil {
		return nil, fmt.Errorf("fetching repos: %w", err)
	}
	var filtered []Repo
	for _, r := range allRepos {
		if r.Archived || r.Fork {
			continue
		}
		if r.Owner.Login != username {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered, nil
}

func fetchOpenPRs(client *api.RESTClient, repoFullName string) ([]PR, error) {
	prs, err := paginateGet[PR](client, fmt.Sprintf("repos/%s/pulls?state=open&per_page=100", repoFullName))
	if err != nil {
		return nil, err
	}
	return prs, nil
}

func fetchOpenIssues(client *api.RESTClient, repoFullName string) ([]Issue, error) {
	issues, err := paginateGet[Issue](client, fmt.Sprintf("repos/%s/issues?state=open&per_page=100", repoFullName))
	if err != nil {
		return nil, err
	}
	// Filter out PRs (the issues endpoint returns PRs too)
	var filtered []Issue
	for _, iss := range issues {
		if iss.PullRequestLinks == nil {
			filtered = append(filtered, iss)
		}
	}
	return filtered, nil
}

func fetchDependabotAlerts(client *api.RESTClient, repoFullName string) ([]DependabotAlert, error) {
	var alerts []DependabotAlert
	err := client.Get(fmt.Sprintf("repos/%s/dependabot/alerts?state=open&per_page=100", repoFullName), &alerts)
	if err != nil {
		var httpErr *api.HTTPError
		if errors.As(err, &httpErr) {
			// 403/404 means dependabot alerts not enabled — silent skip
			if httpErr.StatusCode == 403 || httpErr.StatusCode == 404 {
				return nil, nil
			}
		}
		return nil, err
	}
	return alerts, nil
}

func collectReports(client *api.RESTClient, repos []Repo) []RepoReport {
	reports := make([]RepoReport, len(repos))
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup

	for i, repo := range repos {
		wg.Add(1)
		go func(i int, repo Repo) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			report := RepoReport{
				FullName: repo.FullName,
				Stars:    repo.StargazersCount,
				Forks:    repo.ForksCount,
			}

			if prs, err := fetchOpenPRs(client, repo.FullName); err != nil {
				logWarning("fetching PRs for %s: %v", repo.FullName, err)
			} else {
				report.OpenPRs = prs
			}

			if issues, err := fetchOpenIssues(client, repo.FullName); err != nil {
				logWarning("fetching issues for %s: %v", repo.FullName, err)
			} else {
				report.OpenIssues = issues
			}

			if alerts, err := fetchDependabotAlerts(client, repo.FullName); err != nil {
				logWarning("fetching dependabot alerts for %s: %v", repo.FullName, err)
			} else {
				report.DependabotAlerts = alerts
			}

			reports[i] = report
		}(i, repo)
	}

	wg.Wait()
	return reports
}

func logWarning(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(warningWriter, "warning: %s\n", msg)
}

// warningWriter is set to os.Stderr in main.
var warningWriter io.Writer
