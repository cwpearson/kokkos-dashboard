package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"kokkos-dashboard/github"
)

func marshalAndWrite(v any, name string, perm os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(name), perm)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	log.Printf("write %d to %s", len(data), name)
	err = os.WriteFile(name, data, 0755)
	if err != nil {
		return err
	}
	return nil
}

func fetch(config Config) {
	client := github.NewClient(config.GitHubToken)

	// account for the weekend when fetching two days back
	now := time.Now()
	workdaysFound := 0
	daysBack := 0

	for workdaysFound < 2 {
		daysBack++
		checkDate := now.AddDate(0, 0, -daysBack)
		weekday := checkDate.Weekday()

		// Skip Saturday (6) and Sunday (0)
		if weekday != time.Saturday && weekday != time.Sunday {
			workdaysFound++
		}
	}

	since := now.AddDate(0, 0, -daysBack)
	log.Println("fetching since", since)

	// retrieve data
	for _, repo := range config.Repositories {

		repoOutputDir := filepath.Join(config.FetchDir, repo.Owner, repo.Name)

		log.Printf("Fetching issues for %s/%s...", repo.Owner, repo.Name)
		issues, err := client.GetRecentIssues(repo.Owner, repo.Name, since)
		if err != nil {
			log.Printf("get issues error: %v", err)
			continue
		}
		marshalAndWrite(issues, filepath.Join(repoOutputDir, "issues.json"), 0755)

		log.Printf("Fetching PRs for %s/%s...", repo.Owner, repo.Name)
		prs, err := client.GetRecentPullRequests(repo.Owner, repo.Name, since)
		if err != nil {
			log.Printf("Error fetching PRs: %v", err)
			continue
		}
		marshalAndWrite(prs, filepath.Join(repoOutputDir, "prs.json"), 0755)

		issuesOutputDir := filepath.Join(repoOutputDir, "issues")
		for _, issue := range issues {

			issueOutputDir := filepath.Join(issuesOutputDir, fmt.Sprintf("%d", issue.Number))

			comments, err := client.GetIssueComments(repo.Owner, repo.Name, issue.Number, since)
			if err != nil {
				log.Printf("Error fetching PRs: %v", err)
				continue
			}
			marshalAndWrite(comments, filepath.Join(issueOutputDir, "comments.json"), 0755)

			events, err := client.GetIssueEvents(repo.Owner, repo.Name, issue.Number)
			if err != nil {
				log.Printf("event fetch error: %v", err)
				continue
			}
			marshalAndWrite(events, filepath.Join(issueOutputDir, "events.json"), 0755)
		}

		prsOutputDir := filepath.Join(repoOutputDir, "prs")
		for _, pr := range prs {
			prOutputDir := filepath.Join(prsOutputDir, fmt.Sprintf("%d", pr.Number))
			comments, err := client.GetIssueComments(repo.Owner, repo.Name, pr.Number, since)
			if err != nil {
				log.Printf("Error fetching PRs: %v", err)
				continue
			}
			marshalAndWrite(comments, filepath.Join(prOutputDir, "comments.json"), 0755)

			events, err := client.GetIssueEvents(repo.Owner, repo.Name, pr.Number)
			if err != nil {
				log.Printf("event fetch error: %v", err)
				continue
			}
			marshalAndWrite(events, filepath.Join(prOutputDir, "events.json"), 0755)
		}

	}
}
