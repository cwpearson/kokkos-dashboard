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

func fetch(config Config) error {
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
			return err
		}

		// skip repo with no activity
		if len(issues) == 0 {
			continue
		}

		if err := marshalAndWrite(issues, filepath.Join(repoOutputDir, "issues.json"), 0755); err != nil {
			return err
		}

		issuesOutputDir := filepath.Join(repoOutputDir, "issues")
		for _, issue := range issues {

			issueOutputDir := filepath.Join(issuesOutputDir, fmt.Sprintf("%d", issue.Number))

			comments, err := client.GetIssueComments(repo.Owner, repo.Name, issue.Number, config.Since)
			if err != nil {
				return err
			}
			if err := marshalAndWrite(comments, filepath.Join(issueOutputDir, "comments.json"), 0755); err != nil {
				return err
			}

			events, err := client.GetIssueEvents(repo.Owner, repo.Name, issue.Number)
			if err != nil {
				return err
			}
			if err := marshalAndWrite(events, filepath.Join(issueOutputDir, "events.json"), 0755); err != nil {
				return err
			}
		}
	}
	return nil
}
