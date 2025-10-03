package github

import (
	"encoding/json"
	"fmt"
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
	Draft     bool       `json:"draft"`
	Merged    bool       `json:"merged"`
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

// PullRequestCommit represents a commit in a pull request
type PullRequestCommit struct {
	SHA    string `json:"sha"`
	NodeID string `json:"node_id"`
	Commit struct {
		Author struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"author"`
		Committer struct {
			Name  string    `json:"name"`
			Email string    `json:"email"`
			Date  time.Time `json:"date"`
		} `json:"committer"`
		Message string `json:"message"`
		Tree    struct {
			SHA string `json:"sha"`
			URL string `json:"url"`
		} `json:"tree"`
		URL          string `json:"url"`
		CommentCount int    `json:"comment_count"`
		Verification struct {
			Verified  bool   `json:"verified"`
			Reason    string `json:"reason"`
			Signature string `json:"signature"`
			Payload   string `json:"payload"`
		} `json:"verification"`
	} `json:"commit"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	CommentsURL string `json:"comments_url"`
	Author      *User  `json:"author"`
	Committer   *User  `json:"committer"`
	Parents     []struct {
		SHA     string `json:"sha"`
		URL     string `json:"url"`
		HTMLURL string `json:"html_url"`
	} `json:"parents"`
}

// PullRequestReview represents a review on a pull request
type PullRequestReview struct {
	ID                int64      `json:"id"`
	NodeID            string     `json:"node_id"`
	User              *User      `json:"user"`
	Body              string     `json:"body"`
	State             string     `json:"state"` // APPROVED, CHANGES_REQUESTED, COMMENTED, DISMISSED, PENDING
	HTMLURL           string     `json:"html_url"`
	PullRequestURL    string     `json:"pull_request_url"`
	SubmittedAt       *time.Time `json:"submitted_at"`
	CommitID          string     `json:"commit_id"`
	AuthorAssociation string     `json:"author_association"`
	Links             struct {
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
		PullRequest struct {
			Href string `json:"href"`
		} `json:"pull_request"`
	} `json:"_links"`
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
	req.Header.Set("User-Agent", "cwpearson/kokkos-dashboard")

	resp, err := c.rlClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	return issues, nil
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
	req.Header.Set("User-Agent", "cwpearson/kokkos-dashboard")

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
		req.Header.Set("User-Agent", "cwpearson/kokkos-dashboard")

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

// GetPullRequestCommits retrieves all commits for a specific pull request
func (c *Client) GetPullRequestCommits(owner, repo string, pullNumber int) ([]PullRequestCommit, error) {
	var allCommits []PullRequestCommit
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/commits?page=%d&per_page=%d",
			c.baseURL, owner, repo, pullNumber, page, perPage)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "cwpearson/kokkos-dashboard")

		resp, err := c.rlClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
		}

		var commits []PullRequestCommit
		if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
			return nil, err
		}

		allCommits = append(allCommits, commits...)

		// Check if there are more pages
		if len(commits) < perPage {
			break
		}
		page++
	}

	return allCommits, nil
}

// GetPullRequest retrieves a specific pull request by its number
func (c *Client) GetPullRequest(owner, repo string, pullNumber int) (*PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d",
		c.baseURL, owner, repo, pullNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "cwpearson/kokkos-dashboard")

	resp, err := c.rlClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

// GetPullRequestReviews retrieves all reviews for a specific pull request
func (c *Client) GetPullRequestReviews(owner, repo string, pullNumber int) ([]PullRequestReview, error) {
	var allReviews []PullRequestReview
	page := 1
	perPage := 100

	for {
		url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/reviews?page=%d&per_page=%d",
			c.baseURL, owner, repo, pullNumber, page, perPage)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "cwpearson/kokkos-dashboard")

		resp, err := c.rlClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
		}

		var reviews []PullRequestReview
		if err := json.NewDecoder(resp.Body).Decode(&reviews); err != nil {
			return nil, err
		}

		allReviews = append(allReviews, reviews...)

		// Check if there are more pages
		if len(reviews) < perPage {
			break
		}
		page++
	}

	return allReviews, nil
}
