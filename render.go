package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"kokkos-dashboard/github"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// Helper structures
type RepoData struct {
	Owner  string
	Repo   string
	Issues []Issue
}

type Issue struct {
	github.Issue

	Comments []*github.IssueComment
	Events   []github.IssueEvent
	Commits  []github.PullRequestCommit
}

// Helper functions
func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func loadIssues(path string) ([]github.Issue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var issues []github.Issue
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

func render(config Config) error {
	// Read all owner directories
	ownerDirs, err := os.ReadDir(config.FetchDir)
	if err != nil {
		return fmt.Errorf("failed to read data directory: %w", err)
	}

	// Map to organize data by org/repo
	repoData := make(map[string]*RepoData)

	for _, ownerDir := range ownerDirs {

		if !ownerDir.IsDir() {
			continue
		}

		ownerName := ownerDir.Name()
		ownerPath := filepath.Join(config.FetchDir, ownerName)

		log.Println("process", ownerPath)

		repoDirs, err := os.ReadDir(ownerPath)
		if err != nil {
			return err
		}

		for _, repoDir := range repoDirs {

			repoName := repoDir.Name()
			repoPath := filepath.Join(ownerPath, repoName)

			log.Println("process", repoPath)

			data := &RepoData{
				Owner:  ownerName,
				Repo:   repoName,
				Issues: []Issue{},
			}

			// Process issues.json
			issues, err := loadIssues(filepath.Join(repoPath, "issues.json"))
			if err != nil {
				log.Printf("Warning: failed to load issues for %s: %v", ownerName, err)
			}

			// process issues subdirectory
			issuesDir := filepath.Join(repoPath, "issues")
			for _, issue := range issues {
				issueDir := filepath.Join(issuesDir, fmt.Sprintf("%d", issue.Number))

				commentsPath := filepath.Join(issueDir, "comments.json")
				eventsPath := filepath.Join(issueDir, "events.json")
				commitsPath := filepath.Join(issueDir, "commits.json")

				issueData := Issue{
					Issue:    issue,
					Comments: []*github.IssueComment{},
					Events:   []github.IssueEvent{},
					Commits:  []github.PullRequestCommit{},
				}

				if fData, err := os.ReadFile(commentsPath); err == nil {
					json.Unmarshal(fData, &issueData.Comments)
				}
				if fData, err := os.ReadFile(eventsPath); err == nil {
					json.Unmarshal(fData, &issueData.Events)
				}
				if fData, err := os.ReadFile(commitsPath); err == nil {
					json.Unmarshal(fData, &issueData.Commits)
				}

				// filter out old events
				filteredEvents := []github.IssueEvent{}
				for _, event := range issueData.Events {
					if !event.CreatedAt.Before(config.Since) {
						filteredEvents = append(filteredEvents, event)
					}
				}
				issueData.Events = filteredEvents

				// filter out old commits
				filteredCommits := []github.PullRequestCommit{}
				for _, commit := range issueData.Commits {
					if !commit.Commit.Committer.Date.Before(config.Since) {
						filteredCommits = append(filteredCommits, commit)
					}
				}
				issueData.Commits = filteredCommits

				// render bodies to markdown
				for _, comment := range issueData.Comments {
					comment.Body = string(mdToHTML([]byte(comment.Body)))
				}

				data.Issues = append(data.Issues, issueData)
			}

			repoData[fmt.Sprintf("%s/%s", ownerName, repoName)] = data

		}
	}
	// Render the organized data
	return renderRepoData(repoData, config)
}

func renderRepoData(repoData map[string]*RepoData, config Config) error {
	// Sort repos for consistent output
	var repoKeys []string
	for key := range repoData {
		repoKeys = append(repoKeys, key)
	}
	sort.Strings(repoKeys)

	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
	}).ParseGlob("templates/*.html"))

	os.MkdirAll(config.OutputDir, 0755)

	outputFile, err := os.Create(filepath.Join(config.OutputDir, "index.html"))
	if err != nil {
		log.Fatal("Error creating output file:", err)
	}
	defer outputFile.Close()

	type NavRepo struct {
		URL  string
		Name string
	}
	navRepos := []NavRepo{{"", "all"}}
	for _, repo := range repoData {
		navRepos = append(navRepos, NavRepo{fmt.Sprintf("%s/%s", repo.Owner, repo.Repo), repo.Repo})
	}

	// Execute the template with data
	err = tmpl.ExecuteTemplate(outputFile, "index.html", map[string]any{
		"Repos":       repoData,
		"CurrentYear": time.Now().Year(),
		"BuildDate":   time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"NavRepos":    navRepos,
		"SiteRoot":    config.SiteRoot,
		"Since":       config.Since.UTC().Format("2006-01-02T15:04:05.000Z"),
	})
	if err != nil {
		log.Fatal("Error executing template:", err)
	}

	for _, repo := range repoData {

		outputPath := filepath.Join(config.OutputDir, repo.Owner, repo.Repo, "index.html")
		os.MkdirAll(filepath.Dir(outputPath), 0755)
		outputFile, err := os.Create(outputPath)
		if err != nil {
			log.Fatal("Error creating output file:", err)
		}
		defer outputFile.Close()

		err = tmpl.ExecuteTemplate(outputFile, "repo.html", map[string]any{
			"Repo":        repo,
			"CurrentYear": time.Now().Year(),
			"BuildDate":   time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
			"NavRepos":    navRepos,
			"SiteRoot":    config.SiteRoot,
			"Since":       config.Since.UTC().Format("2006-01-02T15:04:05.000Z"),
		})
		if err != nil {
			log.Fatal("Error executing template:", err)
		}

	}

	outputStaticDir := filepath.Join(config.OutputDir, "static")

	log.Println("remove", outputStaticDir)
	os.RemoveAll(outputStaticDir)

	log.Println("static ->", outputStaticDir)
	if err := os.CopyFS(outputStaticDir, os.DirFS("static")); err != nil {
		log.Fatal(err)
	}

	return nil
}
