package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
)

// newTestClient creates a Client pointing at the given test server.
func newTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	return &Client{
		http:    http.DefaultClient,
		baseURL: serverURL,
		token:   "sk_test_token_12345",
	}
}

func TestClient_Get_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/projects" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"object":"list","items":[]}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	data, err := client.Get("/projects", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if result["object"] != "list" {
		t.Errorf("object = %v, want 'list'", result["object"])
	}
}

func TestClient_Get_WithQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("app_id") != "app123" {
			t.Errorf("app_id = %q, want %q", r.URL.Query().Get("app_id"), "app123")
		}
		if r.URL.Query().Get("expand") != "items" {
			t.Errorf("expand = %q, want %q", r.URL.Query().Get("expand"), "items")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	q := make(map[string][]string)
	q["app_id"] = []string{"app123"}
	q["expand"] = []string{"items"}
	_, err := client.Get("/test", q)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
}

func TestClient_Post_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll: %v", err)
		}

		var reqBody map[string]string
		if err := json.Unmarshal(body, &reqBody); err != nil {
			t.Fatalf("Unmarshal request body: %v", err)
		}
		if reqBody["name"] != "test-project" {
			t.Errorf("name = %q, want %q", reqBody["name"], "test-project")
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"object":"project","id":"proj_123","name":"test-project"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	data, err := client.Post("/projects", map[string]string{"name": "test-project"})
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if result["id"] != "proj_123" {
		t.Errorf("id = %q, want %q", result["id"], "proj_123")
	}
}

func TestClient_Post_NilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// With nil body, Content-Type should not be set
		ct := r.Header.Get("Content-Type")
		if ct == "application/json" {
			t.Errorf("Content-Type should not be application/json for nil body")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Post("/action", nil)
	if err != nil {
		t.Fatalf("Post(nil body) error: %v", err)
	}
}

func TestClient_Delete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/projects/proj_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"object":"deleted","id":"proj_123"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	data, err := client.Delete("/projects/proj_123")
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if result["id"] != "proj_123" {
		t.Errorf("id = %q, want %q", result["id"], "proj_123")
	}
}

func TestClient_AuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer sk_test_token_12345" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer sk_test_token_12345")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Get("/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
}

func TestClient_UserAgentHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != UserAgent {
			t.Errorf("User-Agent = %q, want %q", ua, UserAgent)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Get("/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
}

func TestClient_CustomToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer sk_custom_key" {
			t.Errorf("Authorization = %q, want Bearer sk_custom_key", auth)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := &Client{
		http:    http.DefaultClient,
		baseURL: server.URL,
		token:   "sk_custom_key",
	}
	_, err := client.Get("/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
}

func TestClient_ErrorResponse_ParsesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"object": "error",
			"type": "parameter_error",
			"message": "Invalid parameter: name is required",
			"param": "name",
			"doc_url": "https://docs.revenuecat.com",
			"retryable": false
		}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T: %v", err, err)
	}
	if apiErr.Type != "parameter_error" {
		t.Errorf("Type = %q, want %q", apiErr.Type, "parameter_error")
	}
	if apiErr.Message != "Invalid parameter: name is required" {
		t.Errorf("Message = %q, want %q", apiErr.Message, "Invalid parameter: name is required")
	}
	if apiErr.Param != "name" {
		t.Errorf("Param = %q, want %q", apiErr.Param, "name")
	}
	if apiErr.DocURL != "https://docs.revenuecat.com" {
		t.Errorf("DocURL = %q, want %q", apiErr.DocURL, "https://docs.revenuecat.com")
	}
	if apiErr.Retryable {
		t.Error("Retryable should be false")
	}
}

func TestClient_ErrorResponse_NonJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Should be a plain error, not *Error
	if _, ok := err.(*Error); ok {
		t.Error("expected plain error for non-JSON response, got *Error")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should contain status code 500, got: %v", err)
	}
}

func TestClient_ErrorResponse_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{
			"object": "error",
			"type": "resource_not_found",
			"message": "Project not found",
			"retryable": false
		}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Get("/projects/nonexistent", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Type != "resource_not_found" {
		t.Errorf("Type = %q, want %q", apiErr.Type, "resource_not_found")
	}
}

func TestClient_ErrorResponse_401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{
			"object": "error",
			"type": "authentication_error",
			"message": "Invalid API key",
			"retryable": false
		}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Type != "authentication_error" {
		t.Errorf("Type = %q, want %q", apiErr.Type, "authentication_error")
	}
}

func TestClient_ErrorResponse_429_NonRetryable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{
			"object": "error",
			"type": "rate_limit_error",
			"message": "Rate limit exceeded",
			"retryable": false
		}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}
	if apiErr.Type != "rate_limit_error" {
		t.Errorf("Type = %q, want %q", apiErr.Type, "rate_limit_error")
	}
}

func TestClient_Retry_OnRetryableError(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{
				"object": "error",
				"type": "server_error",
				"message": "Service temporarily unavailable",
				"retryable": true,
				"backoff_ms": 10
			}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	data, err := client.Get("/test", nil)
	if err != nil {
		t.Fatalf("Get() error after retry: %v", err)
	}

	got := atomic.LoadInt32(&attempts)
	if got != 3 {
		t.Errorf("attempts = %d, want 3", got)
	}

	var result map[string]bool
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !result["ok"] {
		t.Error("expected ok=true")
	}
}

func TestClient_Retry_ExhaustedReturnsError(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{
			"object": "error",
			"type": "server_error",
			"message": "Persistent failure",
			"retryable": true,
			"backoff_ms": 10
		}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error after exhausted retries, got nil")
	}

	got := atomic.LoadInt32(&attempts)
	if got != int32(MaxRetries) {
		t.Errorf("attempts = %d, want %d", got, MaxRetries)
	}
}

func TestClient_NoRetry_OnNonRetryableError(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"object": "error",
			"type": "parameter_error",
			"message": "Bad request",
			"retryable": false
		}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	got := atomic.LoadInt32(&attempts)
	if got != 1 {
		t.Errorf("attempts = %d, want 1 (should not retry)", got)
	}
}

func TestClient_Retry_WithCustomBackoff(t *testing.T) {
	var attempts int32
	backoffMs := 50

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{
				"object": "error",
				"type": "server_error",
				"message": "Retry me",
				"retryable": true,
				"backoff_ms": ` + intToJSON(backoffMs) + `
			}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Get("/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	got := atomic.LoadInt32(&attempts)
	if got != 2 {
		t.Errorf("attempts = %d, want 2", got)
	}
}

func TestClient_Error_String(t *testing.T) {
	tests := []struct {
		name string
		err  Error
		want string
	}{
		{
			name: "basic error",
			err: Error{
				Type:    "parameter_error",
				Message: "Name is required",
			},
			want: "parameter_error: Name is required",
		},
		{
			name: "error with doc URL",
			err: Error{
				Type:    "invalid_request",
				Message: "Bad field",
				DocURL:  "https://docs.example.com/errors",
			},
			want: "invalid_request: Bad field\n  See: https://docs.example.com/errors",
		},
		{
			name: "error without doc URL",
			err: Error{
				Type:    "server_error",
				Message: "Internal error",
			},
			want: "server_error: Internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClient_Get_200Range(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"200 OK", 200},
		{"201 Created", 201},
		{"202 Accepted", 202},
		{"204 No Content", 204},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"status":"ok"}`))
			}))
			defer server.Close()

			client := newTestClient(t, server.URL)
			data, err := client.Get("/test", nil)
			if err != nil {
				t.Fatalf("Get() error for status %d: %v", tt.statusCode, err)
			}
			// 204 may have empty body in real APIs; test server always writes a body
			_ = data
		})
	}
}

func TestClient_Post_WithComplexBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var parsed map[string]any
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if parsed["name"] != "test" {
			t.Errorf("name = %v, want test", parsed["name"])
		}
		tags, ok := parsed["tags"].([]any)
		if !ok {
			t.Fatalf("tags not an array: %T", parsed["tags"])
		}
		if len(tags) != 2 {
			t.Errorf("tags length = %d, want 2", len(tags))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Post("/test", map[string]any{
		"name": "test",
		"tags": []string{"a", "b"},
	})
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}
}

func TestClient_ConcurrentRequests(t *testing.T) {
	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requestCount, 1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := client.Get("/test", nil)
			done <- err
		}()
	}

	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			t.Errorf("concurrent request %d error: %v", i, err)
		}
	}

	got := atomic.LoadInt32(&requestCount)
	if got != 10 {
		t.Errorf("request count = %d, want 10", got)
	}
}

func TestPaginateAll_ResolvesNextPageURLs(t *testing.T) {
	tests := []struct {
		name     string
		nextPage func(serverURL string) string
	}{
		{
			name: "absolute URL",
			nextPage: func(serverURL string) string {
				return serverURL + "/v2/projects?starting_after=proj_1"
			},
		},
		{
			name: "path including API version",
			nextPage: func(_ string) string {
				return "/v2/projects?starting_after=proj_1"
			},
		},
		{
			name: "bare resource path",
			nextPage: func(_ string) string {
				return "projects?starting_after=proj_1"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestCount int32
			var nextPage string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/projects" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}

				switch atomic.AddInt32(&requestCount, 1) {
				case 1:
					if got := r.URL.Query().Get("limit"); got != "1" {
						t.Errorf("limit = %q, want 1", got)
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"object":"list","items":[{"object":"project","id":"proj_1","name":"First"}],"next_page":` + stringLiteral(nextPage) + `}`))
				case 2:
					if got := r.URL.Query().Get("starting_after"); got != "proj_1" {
						t.Errorf("starting_after = %q, want proj_1", got)
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"object":"list","items":[{"object":"project","id":"proj_2","name":"Second"}],"next_page":null}`))
				default:
					t.Errorf("unexpected extra request %s", r.URL.String())
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			defer server.Close()

			nextPage = tt.nextPage(server.URL)
			client := newTestClient(t, server.URL+"/v2")
			query := url.Values{}
			query.Set("limit", "1")
			items, err := PaginateAll[Project](client, "/projects", query)
			if err != nil {
				t.Fatalf("PaginateAll() error: %v", err)
			}
			if len(items) != 2 {
				t.Fatalf("items count = %d, want 2", len(items))
			}
			if got := atomic.LoadInt32(&requestCount); got != 2 {
				t.Errorf("request count = %d, want 2", got)
			}
		})
	}
}

func TestClient_PathEncoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	// Paths with special characters should work
	paths := []string{
		"/projects/proj_abc123",
		"/projects/proj-with-dashes",
		"/projects/proj_123/products/prod_456",
	}

	for _, path := range paths {
		_, err := client.Get(path, nil)
		if err != nil {
			t.Errorf("Get(%q) error: %v", path, err)
		}
	}
}

func TestClient_EmptyResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Empty body
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	data, err := client.Get("/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty response, got %q", string(data))
	}
}

func TestClient_LargeResponse(t *testing.T) {
	largePayload := strings.Repeat(`{"item":"data"},`, 1000)
	largePayload = `[` + largePayload[:len(largePayload)-1] + `]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largePayload))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	data, err := client.Get("/test", nil)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	var items []map[string]string
	if err := json.Unmarshal(data, &items); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(items) != 1000 {
		t.Errorf("items count = %d, want 1000", len(items))
	}
}

func TestConstants(t *testing.T) {
	if BaseURL != "https://api.revenuecat.com/v2" {
		t.Errorf("BaseURL = %q, want %q", BaseURL, "https://api.revenuecat.com/v2")
	}
	if UserAgent == "" {
		t.Error("UserAgent should not be empty")
	}
	if !strings.HasPrefix(UserAgent, "rc-cli/") {
		t.Errorf("UserAgent = %q, should start with 'rc-cli/'", UserAgent)
	}
	if MaxRetries < 1 {
		t.Errorf("MaxRetries = %d, should be >= 1", MaxRetries)
	}
}

func TestClient_ErrorResponse_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(""))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error for 400 with empty body")
	}
}

func TestClient_ErrorResponse_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{invalid json"))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)
	_, err := client.Get("/test", nil)
	if err == nil {
		t.Fatal("expected error for 400 with invalid JSON")
	}

	// Should be a plain error, not *Error
	if _, ok := err.(*Error); ok {
		t.Error("expected plain error for invalid JSON, got *Error")
	}
}

func TestListResponse_Generic(t *testing.T) {
	jsonData := `{
		"object": "list",
		"items": [
			{"object": "project", "id": "proj_1", "name": "First"},
			{"object": "project", "id": "proj_2", "name": "Second"}
		],
		"next_page": "cursor_abc",
		"url": "/projects"
	}`

	var resp ListResponse[Project]
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if resp.Object != "list" {
		t.Errorf("Object = %q, want 'list'", resp.Object)
	}
	if len(resp.Items) != 2 {
		t.Fatalf("Items count = %d, want 2", len(resp.Items))
	}
	if resp.Items[0].ID != "proj_1" {
		t.Errorf("Items[0].ID = %q, want %q", resp.Items[0].ID, "proj_1")
	}
	if resp.NextPage == nil || *resp.NextPage != "cursor_abc" {
		t.Errorf("NextPage = %v, want 'cursor_abc'", resp.NextPage)
	}
	if resp.URL != "/projects" {
		t.Errorf("URL = %q, want %q", resp.URL, "/projects")
	}
}

func TestListResponse_Empty(t *testing.T) {
	jsonData := `{"object": "list", "items": [], "url": "/test"}`

	var resp ListResponse[Project]
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(resp.Items) != 0 {
		t.Errorf("Items count = %d, want 0", len(resp.Items))
	}
	if resp.NextPage != nil {
		t.Errorf("NextPage = %v, want nil", resp.NextPage)
	}
}

func intToJSON(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func stringLiteral(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
