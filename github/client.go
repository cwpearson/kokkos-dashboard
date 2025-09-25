package github

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"kokkos-dashboard/ratelimit"
)

type Client struct {
	token    string
	rlClient *ratelimit.Client
	baseURL  string
}

type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	HTMLURL   string    `json:"html_url"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	PullRequest *struct{} `json:"pull_request,omitempty"` // to filter out PRs
}

type PullRequest struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	State     string     `json:"state"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	MergedAt  *time.Time `json:"merged_at"`
	HTMLURL   string     `json:"html_url"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
}

type IssueComment struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	HTMLURL   string    `json:"html_url"`
	User      struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
	IssueURL string `json:"issue_url"`
}

// User represents a GitHub user
type User struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	NodeID    string `json:"node_id"`
	AvatarURL string `json:"avatar_url"`
	HTMLURL   string `json:"html_url"`
	Type      string `json:"type"`
}

type IssueEvent struct {
	ID        int64     `json:"id"`
	NodeID    string    `json:"node_id"`
	URL       string    `json:"url"`
	Actor     *User     `json:"actor"`
	Event     string    `json:"event"`
	CommitID  *string   `json:"commit_id"`
	CommitURL *string   `json:"commit_url"`
	CreatedAt time.Time `json:"created_at"`

	// // Event-specific fields
	// Label             *Label           `json:"label,omitempty"`
	// Assignee          *User            `json:"assignee,omitempty"`
	// Assigner          *User            `json:"assigner,omitempty"`
	// Milestone         *Milestone       `json:"milestone,omitempty"`
	// Rename            *Rename          `json:"rename,omitempty"`
	// ReviewRequester   *User            `json:"review_requester,omitempty"`
	// RequestedReviewer *User            `json:"requested_reviewer,omitempty"`
	// RequestedTeam     *Team            `json:"requested_team,omitempty"`
	// DismissedReview   *DismissedReview `json:"dismissed_review,omitempty"`
	// LockReason        *string          `json:"lock_reason,omitempty"`
	// ProjectCard       *ProjectCard     `json:"project_card,omitempty"`

	// // Issue reference (for some events)
	// Issue *Issue `json:"issue,omitempty"`
}

func NewClient(token string) *Client {

	return &Client{
		token: token,
		rlClient: ratelimit.NewRateLimitedClientWithHTTPClient(
			&http.Client{Timeout: 30 * time.Second},
			time.Hour/5000, // 5k requests per hour
		),
		baseURL: "https://api.github.com",
	}
}

func (c *Client) GetRecentIssues(owner, repo string, since time.Time) ([]Issue, error) {

	// TODO: since should be in ISO 8601
	url := fmt.Sprintf("%s/repos/%s/%s/issues?state=all&sort=updated&since=%s&per_page=100",
		c.baseURL, owner, repo, since.Format(time.RFC3339))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.rlClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	// Filter out pull requests (GitHub API returns PRs as issues too)
	var filteredIssues []Issue
	for _, issue := range issues {
		if issue.PullRequest == nil {
			filteredIssues = append(filteredIssues, issue)
		} else {
			log.Println("issue", issue.Number, "was actually a PR")
		}
	}

	return filteredIssues, nil
}

func (c *Client) GetRecentPullRequests(owner, repo string, since time.Time) ([]PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls?state=all&sort=updated&direction=desc&since=%s&per_page=100",
		c.baseURL, owner, repo, since.Format(time.RFC3339))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.rlClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var prs []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, err
	}

	// Filter PRs updated since the given time
	var recentPRs []PullRequest
	for _, pr := range prs {
		if pr.UpdatedAt.After(since) {
			recentPRs = append(recentPRs, pr)
		} else {
			break // Since they're sorted by updated time
		}
	}

	return recentPRs, nil
}

// GetIssueComments retrieves all comments for a specific issue since a given timestamp
func (c *Client) GetIssueComments(owner, repo string, issueNumber int, since time.Time) ([]IssueComment, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments?sort=updated&since=%s&per_page=100",
		c.baseURL, owner, repo, issueNumber, since.Format(time.RFC3339))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.rlClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var comments []IssueComment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// GetIssueEvents retrieves all events for a specific issue
func (c *Client) GetIssueEvents(owner, repo string, issueNumber int) ([]IssueEvent, error) {
	var allEvents []IssueEvent
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/events?page=%d&per_page=%d",
			c.baseURL, owner, repo, issueNumber, page, perPage)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := c.rlClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
		}

		var events []IssueEvent
		if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
			return nil, err
		}

		allEvents = append(allEvents, events...)

		// Check if there are more pages
		if len(events) < perPage {
			break
		}
		page++
	}

	return allEvents, nil
}
