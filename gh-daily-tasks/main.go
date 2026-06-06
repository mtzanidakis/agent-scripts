package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

func main() {
	warningWriter = os.Stderr

	db, err := openDB()
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer func() { _ = db.Close() }()

	client, err := api.DefaultRESTClient()
	if err != nil {
		log.Fatalf("creating GitHub client (run `gh auth login`): %v", err)
	}

	username, err := fetchAuthenticatedUser(client)
	if err != nil {
		log.Fatal(err)
	}

	repos, err := fetchAllRepos(client, username)
	if err != nil {
		log.Fatal(err)
	}

	reports := collectReports(client, repos)

	// Get previous snapshots BEFORE saving new ones
	repoNames := make([]string, len(reports))
	for i, r := range reports {
		repoNames[i] = r.FullName
	}
	previous, err := getPreviousSnapshots(db, repoNames)
	if err != nil {
		log.Fatalf("reading previous snapshots: %v", err)
	}

	deltas := computeDeltas(reports, previous)

	// Save current snapshots
	if err := saveSnapshots(db, reports); err != nil {
		log.Fatalf("saving snapshots: %v", err)
	}

	// Build output
	output := buildOutput(reports, deltas)

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("marshaling JSON: %v", err)
	}
	fmt.Println(string(data))
}

func buildOutput(reports []RepoReport, deltas map[string][2]int) OutputReport {
	output := OutputReport{
		Date:       time.Now().Format("2006-01-02"),
		TotalRepos: len(reports),
		Summary: OutputSummary{
			AlertsBySeverity: make(map[string]int),
		},
	}

	for _, r := range reports {
		d := deltas[r.FullName]
		starsDelta, forksDelta := d[0], d[1]

		// Skip repos with no activity
		if len(r.OpenPRs) == 0 && len(r.OpenIssues) == 0 && len(r.DependabotAlerts) == 0 && starsDelta == 0 && forksDelta == 0 {
			continue
		}

		repo := OutputRepo{
			Name:  r.FullName,
			Stars: Delta{Current: r.Stars, Delta: starsDelta},
			Forks: Delta{Current: r.Forks, Delta: forksDelta},
		}

		for _, pr := range r.OpenPRs {
			repo.OpenPRs = append(repo.OpenPRs, OutputPR{
				Number:    pr.Number,
				Title:     pr.Title,
				Author:    pr.User.Login,
				URL:       pr.HTMLURL,
				CreatedAt: pr.CreatedAt.Format(time.RFC3339),
				Draft:     pr.Draft,
			})
		}

		for _, iss := range r.OpenIssues {
			var labels []string
			for _, l := range iss.Labels {
				labels = append(labels, l.Name)
			}
			repo.OpenIssues = append(repo.OpenIssues, OutputIssue{
				Number:    iss.Number,
				Title:     iss.Title,
				Author:    iss.User.Login,
				URL:       iss.HTMLURL,
				CreatedAt: iss.CreatedAt.Format(time.RFC3339),
				Labels:    labels,
			})
		}

		for _, alert := range r.DependabotAlerts {
			severity := alert.SecurityAdvisory.Severity
			repo.DependabotAlerts = append(repo.DependabotAlerts, OutputAlert{
				Number:   alert.Number,
				Severity: severity,
				Summary:  alert.SecurityAdvisory.Summary,
				URL:      alert.HTMLURL,
			})
			output.Summary.AlertsBySeverity[severity]++
		}

		output.Summary.TotalOpenPRs += len(r.OpenPRs)
		output.Summary.TotalOpenIssues += len(r.OpenIssues)
		output.Summary.TotalAlerts += len(r.DependabotAlerts)
		output.Summary.NetStarsDelta += starsDelta
		output.Summary.NetForksDelta += forksDelta

		output.Repos = append(output.Repos, repo)
	}

	return output
}
