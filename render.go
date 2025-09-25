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
	PRs    []PR
}

type Issue struct {
	github.Issue

	Comments []*github.IssueComment
	Events   []github.IssueEvent
}

type PR struct {
	github.PullRequest

	Comments []*github.IssueComment
	Events   []github.IssueEvent
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

func loadPullRequests(path string) ([]github.PullRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var prs []github.PullRequest
	if err := json.Unmarshal(data, &prs); err != nil {
		return nil, err
	}
	return prs, nil
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
				PRs:    []PR{},
			}

			// Process issues.json
			issues, err := loadIssues(filepath.Join(repoPath, "issues.json"))
			if err != nil {
				log.Printf("Warning: failed to load issues for %s: %v", ownerName, err)
			}

			// Process prs.json
			prs, err := loadPullRequests(filepath.Join(repoPath, "prs.json"))
			if err != nil {
				log.Printf("Warning: failed to load PRs for %s: %v", ownerName, err)
			}

			// process issues subdirectory
			issuesDir := filepath.Join(repoPath, "issues")
			for _, issue := range issues {
				issueDir := filepath.Join(issuesDir, fmt.Sprintf("%d", issue.Number))

				commentsPath := filepath.Join(issueDir, "comments.json")
				eventsPath := filepath.Join(issueDir, "events.json")

				issueData := Issue{
					Issue:    issue,
					Comments: []*github.IssueComment{},
					Events:   []github.IssueEvent{},
				}

				if fData, err := os.ReadFile(commentsPath); err == nil {
					json.Unmarshal(fData, &issueData.Comments)
				}
				if fData, err := os.ReadFile(eventsPath); err == nil {
					json.Unmarshal(fData, &issueData.Events)
				}

				// render bodies to markdown
				for _, comment := range issueData.Comments {
					comment.Body = string(mdToHTML([]byte(comment.Body)))
				}

				data.Issues = append(data.Issues, issueData)
			}

			// process PRs subdirectory
			prsDir := filepath.Join(repoPath, "prs")
			for _, pr := range prs {
				prDir := filepath.Join(prsDir, fmt.Sprintf("%d", pr.Number))

				commentsPath := filepath.Join(prDir, "comments.json")
				eventsPath := filepath.Join(prDir, "events.json")

				prData := PR{
					PullRequest: pr,
					Comments:    []*github.IssueComment{},
					Events:      []github.IssueEvent{},
				}

				if fData, err := os.ReadFile(commentsPath); err == nil {
					json.Unmarshal(fData, &prData.Comments)
				}
				if fData, err := os.ReadFile(eventsPath); err == nil {
					json.Unmarshal(fData, &prData.Events)
				}

				// render bodies to markdown
				for _, comment := range prData.Comments {
					comment.Body = string(mdToHTML([]byte(comment.Body)))
				}

				data.PRs = append(data.PRs, prData)
			}

			repoData[fmt.Sprintf("%s/%s", ownerName, repoName)] = data

		}
	}
	// Render the organized data
	return renderRepoData(repoData, config)
}

type IssueActivity struct {
	Comments []github.IssueComment
	Events   []github.IssueEvent
}

type PRActivity struct {
	Comments []github.IssueComment
	Events   []github.IssueEvent
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

	// Execute the template with data
	err = tmpl.ExecuteTemplate(outputFile, "index.html", map[string]any{
		"Repos":       repoData,
		"CurrentYear": time.Now().Year(),
		"BuildDate":   time.Now().Format(time.RFC3339),
	})
	if err != nil {
		log.Fatal("Error executing template:", err)
	}

	os.RemoveAll(filepath.Join(config.OutputDir, "static"))
	if err := os.CopyFS(filepath.Join(config.OutputDir, "static"), os.DirFS("static")); err != nil {
		log.Fatal(err)
	}

	return nil
}
