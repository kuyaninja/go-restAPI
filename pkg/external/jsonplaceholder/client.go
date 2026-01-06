package jsonplaceholder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	defaultBaseURL     = "https://jsonplaceholder.typicode.com"
	defaultHTTPTimeout = 10 * time.Second
)

// HTTPClient describes the subset of Fiber's client used by the placeholder client.
type HTTPClient interface {
	Get(url string) *fiber.Agent
}

// Client coordinates calls to the JSONPlaceholder API.
type Client struct {
	baseURL string
	http    HTTPClient
}

// Post mirrors the response from the JSONPlaceholder /posts endpoint.
type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// ErrMissingPostID Sentinel errors returned by the client.
var (
	ErrMissingPostID = errors.New("post id is required")
)

// NewClient builds a placeholder API client using the provided base URL and HTTP client.
// When httpClient is nil, a new http.Client with sane timeouts is used.
func NewClient(baseURL string, httpClient HTTPClient) *Client {
	trimmed := strings.TrimSpace(baseURL)
	trimmed = strings.TrimRight(trimmed, "/")
	if trimmed == "" {
		trimmed = defaultBaseURL
	}

	if httpClient == nil {
		httpClient = &fiber.Client{}
	}

	return &Client{baseURL: trimmed, http: httpClient}
}

// GetPost fetches a post by id from JSONPlaceholder.
func (c *Client) GetPost(ctx context.Context, id string) (Post, error) {
	var post Post
	if strings.TrimSpace(id) == "" {
		return post, ErrMissingPostID
	}

	url := fmt.Sprintf("%s/posts/%s", c.baseURL, strings.TrimSpace(id))
	if err := ctx.Err(); err != nil {
		return post, err
	}

	agent := c.http.Get(url)
	agent.Timeout(resolveTimeout(ctx))
	status, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return post, errs[0]
	}

	if status != fiber.StatusOK {
		snippet := strings.TrimSpace(string(body))
		if len(snippet) > 512 {
			snippet = snippet[:512]
		}
		return post, fmt.Errorf("jsonplaceholder: unexpected status %d: %s", status, snippet)
	}

	if err := json.Unmarshal(body, &post); err != nil {
		return post, err
	}

	return post, nil
}

func resolveTimeout(ctx context.Context) time.Duration {
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining > 0 && remaining < defaultHTTPTimeout {
			return remaining
		}
	}
	return defaultHTTPTimeout
}
