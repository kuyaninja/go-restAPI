package jsonplaceholder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL     = "https://jsonplaceholder.typicode.com"
	defaultHTTPTimeout = 10 * time.Second
)

// HTTPClient describes the subset of http.Client used by the placeholder client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client coordinates calls to the JSONPlaceholder API.
type Client struct {
	baseURL    string
	httpClient HTTPClient
}

// Post mirrors the response from the JSONPlaceholder /posts endpoint.
type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// Sentinel errors returned by the client.
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
		httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	}

	return &Client{baseURL: trimmed, httpClient: httpClient}
}

// GetPost fetches a post by id from JSONPlaceholder.
func (c *Client) GetPost(ctx context.Context, id string) (Post, error) {
	var post Post
	if strings.TrimSpace(id) == "" {
		return post, ErrMissingPostID
	}

	url := fmt.Sprintf("%s/posts/%s", c.baseURL, strings.TrimSpace(id))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return post, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return post, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return post, fmt.Errorf("jsonplaceholder: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(snippet)))
	}

	if err := json.NewDecoder(resp.Body).Decode(&post); err != nil {
		return post, err
	}

	return post, nil
}
