package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/andresdefi/rc/internal/auth"
)

const (
	BaseURL    = "https://api.revenuecat.com/v2"
	UserAgent  = "rc-cli/0.1.0"
	MaxRetries = 3
)

// Client is the RevenueCat API v2 HTTP client.
type Client struct {
	http    *http.Client
	baseURL string
	token   string
}

// NewClient creates a new API client using the stored auth token.
func NewClient() (*Client, error) {
	token, err := auth.GetToken()
	if err != nil {
		return nil, err
	}
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: BaseURL,
		token:   token,
	}, nil
}

// Error represents a RevenueCat API error response.
type Error struct {
	Object    string `json:"object"`
	Type      string `json:"type"`
	Param     string `json:"param,omitempty"`
	DocURL    string `json:"doc_url,omitempty"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	BackoffMs *int   `json:"backoff_ms,omitempty"`
}

func (e *Error) Error() string {
	msg := fmt.Sprintf("%s: %s", e.Type, e.Message)
	if e.DocURL != "" {
		msg += fmt.Sprintf("\n  See: %s", e.DocURL)
	}
	return msg
}

// ListResponse is the generic paginated list envelope.
type ListResponse[T any] struct {
	Object   string  `json:"object"`
	Items    []T     `json:"items"`
	NextPage *string `json:"next_page"`
	URL      string  `json:"url"`
}

// Get performs a GET request.
func (c *Client) Get(path string, query url.Values) ([]byte, error) {
	return c.do("GET", path, query, nil)
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(path string, body any) ([]byte, error) {
	return c.do("POST", path, nil, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) ([]byte, error) {
	return c.do("DELETE", path, nil, nil)
}

func (c *Client) do(method, path string, query url.Values, body any) ([]byte, error) {
	u := c.baseURL + path
	if query != nil && len(query) > 0 {
		u += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	var lastErr error
	for attempt := range MaxRetries {
		req, err := http.NewRequest(method, u, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("User-Agent", UserAgent)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		var apiErr Error
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Message != "" {
			if apiErr.Retryable && attempt < MaxRetries-1 {
				backoff := time.Duration(100*(attempt+1)) * time.Millisecond
				if apiErr.BackoffMs != nil {
					backoff = time.Duration(*apiErr.BackoffMs) * time.Millisecond
				}
				time.Sleep(backoff)
				lastErr = &apiErr
				continue
			}
			return nil, &apiErr
		}

		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", MaxRetries, lastErr)
}
