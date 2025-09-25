package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

	log.Println("fetching since", config.Since)

	log.Println("remove", config.FetchDir)
	os.RemoveAll(config.FetchDir)

	// retrieve data
	for _, repo := range config.Repositories {

		repoOutputDir := filepath.Join(config.FetchDir, repo.Owner, repo.Name)

		log.Printf("Fetching issues for %s/%s...", repo.Owner, repo.Name)
		issues, err := client.GetRecentIssues(repo.Owner, repo.Name, config.Since)
		if err != nil {
			log.Printf("get issues error: %v", err)
			continue
		}
		marshalAndWrite(issues, filepath.Join(repoOutputDir, "issues.json"), 0755)

		issuesOutputDir := filepath.Join(repoOutputDir, "issues")
		for _, issue := range issues {

			issueOutputDir := filepath.Join(issuesOutputDir, fmt.Sprintf("%d", issue.Number))

			comments, err := client.GetIssueComments(repo.Owner, repo.Name, issue.Number, config.Since)
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
	}
}
