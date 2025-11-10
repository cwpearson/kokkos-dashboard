package ratelimit

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Client wraps an HTTP client with per-domain rate limiting
type Client struct {
	client      *http.Client
	minInterval time.Duration
	domainLocks map[string]*domainLimiter
	mu          sync.RWMutex
}

// domainLimiter tracks the last request time for a specific domain
type domainLimiter struct {
	lastRequest time.Time
	mu          sync.Mutex
}

// NewRateLimitedClient creates a new rate-limited HTTP client
func NewRateLimitedClient(minInterval time.Duration) *Client {
	return &Client{
		client:      &http.Client{Timeout: 30 * time.Second},
		minInterval: minInterval,
		domainLocks: make(map[string]*domainLimiter),
	}
}

// NewRateLimitedClientWithHTTPClient creates a new rate-limited client with a custom HTTP client
func NewRateLimitedClientWithHTTPClient(client *http.Client, minInterval time.Duration) *Client {
	return &Client{
		client:      client,
		minInterval: minInterval,
		domainLocks: make(map[string]*domainLimiter),
	}
}

// getDomainLimiter returns or creates a limiter for the given domain
func (c *Client) getDomainLimiter(domain string) *domainLimiter {
	c.mu.RLock()
	limiter, exists := c.domainLocks[domain]
	c.mu.RUnlock()

	if exists {
		return limiter
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := c.domainLocks[domain]; exists {
		return limiter
	}

	limiter = &domainLimiter{}
	c.domainLocks[domain] = limiter
	return limiter
}

// Do executes an HTTP request with rate limiting
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	domain := req.URL.Host
	if domain == "" {
		return nil, fmt.Errorf("request URL has no host")
	}

	limiter := c.getDomainLimiter(domain)

	// Acquire the lock for this domain
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	// Calculate how long to wait
	timeSinceLastRequest := time.Since(limiter.lastRequest)
	if timeSinceLastRequest < c.minInterval {
		waitTime := c.minInterval - timeSinceLastRequest
		log.Println("wait for", waitTime, "for", domain)
		time.Sleep(waitTime)
	}

	// Update last request time
	limiter.lastRequest = time.Now()

	// Execute the request
	var resp *http.Response
	var err error
	for range 3 {
		resp, err = c.client.Do(req)

		if resp.StatusCode < 500 {
			break
		}

		// sleep before retry
		time.Sleep(5 * time.Second)
	}

	// Update last request time
	limiter.lastRequest = time.Now()

	return resp, err
}

// SetMinInterval updates the minimum interval between requests
func (c *Client) SetMinInterval(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.minInterval = interval
}

// ClearDomainHistory removes rate limiting history for a specific domain
func (c *Client) ClearDomainHistory(domain string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.domainLocks, domain)
}

// ClearAllHistory removes all rate limiting history
func (c *Client) ClearAllHistory() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.domainLocks = make(map[string]*domainLimiter)
}
