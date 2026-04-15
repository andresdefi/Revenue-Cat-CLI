package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/andresdefi/rc/internal/auth"
	"github.com/andresdefi/rc/internal/cache"
	"github.com/andresdefi/rc/internal/output"
	"github.com/andresdefi/rc/internal/version"
)

const (
	MaxRetries = 3
)

var UserAgent = "rc-cli/" + version.Version

var BaseURL = "https://api.revenuecat.com/v2"

// DryRun when true prevents POST/DELETE requests from being sent.
var DryRun bool

// CacheEnabled enables local response caching for GET requests.
var CacheEnabled bool

// Client is the RevenueCat API v2 HTTP client.
type Client struct {
	http    *http.Client
	baseURL string
	token   string
}

// NewClient creates a new API client using the stored auth token for the active profile.
func NewClient() (*Client, error) {
	return NewClientForProfile("")
}

// NewClientForProfile creates a new API client using the stored auth token for the given profile.
func NewClientForProfile(profile string) (*Client, error) {
	token, err := auth.GetToken(profile)
	if err != nil {
		return nil, err
	}
	return NewClientWithToken(token), nil
}

// NewClientWithToken creates a new API client using the provided token directly.
func NewClientWithToken(token string) *Client {
	return &Client{
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: BaseURL,
		token:   token,
	}
}

// Error represents a RevenueCat API error response.
type Error struct {
	Object     string `json:"object"`
	Type       string `json:"type"`
	Param      string `json:"param,omitempty"`
	DocURL     string `json:"doc_url,omitempty"`
	Message    string `json:"message"`
	Retryable  bool   `json:"retryable"`
	BackoffMs  *int   `json:"backoff_ms,omitempty"`
	StatusCode int    `json:"-"`
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

// Get performs a GET request. Results are cached when CacheEnabled is true.
func (c *Client) Get(path string, query url.Values) ([]byte, error) {
	cacheKey := path
	if len(query) > 0 {
		cacheKey += "?" + query.Encode()
	}
	if CacheEnabled {
		if cached := cache.Get(cacheKey); cached != nil {
			output.Debug("Cache hit: %s", cacheKey)
			return cached, nil
		}
	}
	data, err := c.do("GET", path, query, nil)
	if err == nil && CacheEnabled {
		cache.Set(cacheKey, data)
	}
	return data, err
}

// Post performs a POST request with a JSON body.
func (c *Client) Post(path string, body any) ([]byte, error) {
	return c.do("POST", path, nil, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(path string) ([]byte, error) {
	return c.do("DELETE", path, nil, nil)
}

// GetFullURL performs a GET using a full URL path (e.g. from next_page).
// The path should already include any query parameters.
func (c *Client) GetFullURL(fullPath string) ([]byte, error) {
	u, err := c.resolvePageURL(fullPath)
	if err != nil {
		return nil, err
	}
	output.Debug("GET %s", u)

	var lastErr error
	for attempt := range MaxRetries {
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("User-Agent", UserAgent)

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			output.Debug("Attempt %d failed: %v", attempt+1, err)
			continue
		}

		respBody, err := readResponseBody(resp)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		var apiErr Error
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Message != "" {
			apiErr.StatusCode = resp.StatusCode
			if apiErr.Retryable && attempt < MaxRetries-1 {
				backoff := retryBackoff(resp, &apiErr, attempt)
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

func (c *Client) resolvePageURL(nextPage string) (string, error) {
	if nextPage == "" {
		return "", fmt.Errorf("empty next page URL")
	}
	if parsed, err := url.Parse(nextPage); err == nil && parsed.IsAbs() {
		return nextPage, nil
	}

	base, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	nextPath := nextPage
	if !strings.HasPrefix(nextPath, "/") {
		nextPath = "/" + nextPath
	}
	if base.Path != "" && strings.HasPrefix(nextPath, base.Path+"/") {
		nextPath = strings.TrimPrefix(nextPath, base.Path)
	}
	return strings.TrimRight(c.baseURL, "/") + nextPath, nil
}

func (c *Client) do(method, path string, query url.Values, body any) ([]byte, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var bodyData []byte
	var bodyReader io.Reader
	if body != nil {
		var err error
		bodyData, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyData)
	}

	output.Debug("%s %s", method, u)
	if len(bodyData) > 0 {
		output.Debug("Request body: %s", string(bodyData))
	}

	if DryRun && method != "GET" {
		output.Debug("[dry-run] Skipping %s %s", method, path)
		fmt.Fprintf(os.Stderr, "[dry-run] Would %s %s\n", method, path)
		if len(bodyData) > 0 {
			var pretty bytes.Buffer
			if json.Indent(&pretty, bodyData, "  ", "  ") == nil {
				fmt.Fprintf(os.Stderr, "  Body:\n  %s\n", pretty.String())
			}
		}
		return []byte(`{}`), nil
	}

	var lastErr error
	for attempt := range MaxRetries {
		if bodyReader != nil {
			bodyReader = bytes.NewReader(bodyData)
		}

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
			output.Debug("Attempt %d failed: %v", attempt+1, err)
			continue
		}

		respBody, err := readResponseBody(resp)
		if err != nil {
			return nil, err
		}

		output.Debug("Response: %d (%d bytes)", resp.StatusCode, len(respBody))
		if output.Verbose && len(respBody) > 0 && len(respBody) < 4096 {
			output.Debug("Response body: %s", string(respBody))
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		var apiErr Error
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Message != "" {
			apiErr.StatusCode = resp.StatusCode
			if apiErr.Retryable && attempt < MaxRetries-1 {
				backoff := retryBackoff(resp, &apiErr, attempt)
				output.Debug("Retryable error, backing off %v", backoff)
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

// retryBackoff determines the backoff duration from the Retry-After header,
// the backoff_ms JSON field, or a default exponential backoff.
func retryBackoff(resp *http.Response, apiErr *Error, attempt int) time.Duration {
	if ra := resp.Header.Get("Retry-After"); ra != "" {
		if seconds, err := strconv.Atoi(ra); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	if apiErr.BackoffMs != nil {
		return time.Duration(*apiErr.BackoffMs) * time.Millisecond
	}
	return time.Duration(100*(attempt+1)) * time.Millisecond
}

func readResponseBody(resp *http.Response) ([]byte, error) {
	respBody, readErr := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("failed to close response body: %w", closeErr)
	}
	return respBody, nil
}

// Paginate fetches all pages from a list endpoint, calling fn for each page.
// fn receives the items from each page and returns (keepGoing, error).
// If fn returns false, pagination stops.
func Paginate[T any](c *Client, path string, query url.Values, fn func(items []T) (bool, error)) error {
	// Build the initial URL with query params.
	initialPath := path
	if len(query) > 0 {
		initialPath += "?" + query.Encode()
	}

	currentPath := initialPath
	isFirst := true

	for {
		var data []byte
		var err error
		if isFirst {
			data, err = c.Get(path, query)
			isFirst = false
		} else {
			// currentPath is a full path from next_page (includes query params)
			data, err = c.GetFullURL(currentPath)
		}
		if err != nil {
			return err
		}

		var resp ListResponse[T]
		if err := json.Unmarshal(data, &resp); err != nil {
			return fmt.Errorf("failed to parse paginated response: %w", err)
		}

		keepGoing, err := fn(resp.Items)
		if err != nil {
			return err
		}
		if !keepGoing {
			return nil
		}

		if resp.NextPage == nil {
			return nil
		}
		currentPath = *resp.NextPage
	}
}

// PaginateAll collects all items across all pages.
func PaginateAll[T any](c *Client, path string, query url.Values) ([]T, error) {
	var all []T
	err := Paginate(c, path, query, func(items []T) (bool, error) {
		all = append(all, items...)
		return true, nil
	})
	return all, err
}
