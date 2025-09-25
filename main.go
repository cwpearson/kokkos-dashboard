package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"kokkos-dashboard/github"
)

type Config struct {
	GitHubToken  string
	Repositories []struct {
		Owner string
		Name  string
	}
	FetchDir  string
	OutputDir string
}

type Activity struct {
}

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
	since := time.Now().AddDate(0, 0, -1) // Last 1 days

	// retrieve data
	for _, repo := range config.Repositories {

		repoOutputDir := filepath.Join(config.FetchDir, repo.Owner, repo.Name)

		log.Printf("Fetching issues for %s/%s...", repo.Owner, repo.Name)
		issues, err := client.GetRecentIssues(repo.Owner, repo.Name, since)
		if err != nil {
			log.Printf("Error fetching issues: %v", err)
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

func main() {
	// Define command-line flags
	fetchFlag := flag.Bool("fetch", false, "Fetch GitHub activity data")
	renderFlag := flag.Bool("render", false, "Render static site from fetched data")
	flag.Parse()

	// Load configuration
	config := Config{
		GitHubToken: os.Getenv("KOKKOS_DASHBOARD_TOKEN"),
		Repositories: []struct {
			Owner string
			Name  string
		}{
			{Owner: "kokkos", Name: "kokkos"},
			{Owner: "kokkos", Name: "kokkos-kernels"},
			{Owner: "kokkos", Name: "kokkos-comm"},
			{Owner: "kokkos", Name: "kokkos-fft"},
			{Owner: "kokkos", Name: "kokkos-core-wiki"},
		},
		FetchDir:  "data/",
		OutputDir: "public/",
	}

	// fetch
	if *fetchFlag {
		fmt.Println("Fetching GitHub activity...")
		fetch(config)
	}

	// render
	if *renderFlag {
		fmt.Println("Rendering static site...")
		render(config)
	}
}
