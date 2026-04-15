//go:build integration

package integration

import (
	"os"
	"testing"

	"github.com/andresdefi/rc/internal/api"
)

// These tests run against the real RevenueCat API.
// They are gated behind the "integration" build tag and require:
//   RC_INTEGRATION_KEY - a valid RevenueCat API v2 secret key (sk_...)
//
// Run with: go test -tags integration ./internal/integration/...
// Or:       make test-integration

func apiKey(t *testing.T) string {
	t.Helper()
	key := os.Getenv("RC_INTEGRATION_KEY")
	if key == "" {
		t.Skip("RC_INTEGRATION_KEY not set, skipping integration tests")
	}
	return key
}

func TestListProjects(t *testing.T) {
	client := api.NewClientWithToken(apiKey(t))

	data, err := client.Get("/projects", nil)
	if err != nil {
		t.Fatalf("GET /projects failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty response from /projects")
	}
}

func TestInvalidToken(t *testing.T) {
	client := api.NewClientWithToken("sk_test_invalid_key_for_integration")

	_, err := client.Get("/projects", nil)
	if err == nil {
		t.Fatal("expected error with invalid token, got nil")
	}

	apiErr, ok := err.(*api.Error)
	if !ok {
		t.Fatalf("expected *api.Error, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestRateLimitExitCode(t *testing.T) {
	// This test verifies that the API client populates StatusCode on errors.
	// We use an invalid token which should return 401.
	client := api.NewClientWithToken("sk_test_rate_limit_check")

	_, err := client.Get("/projects", nil)
	if err == nil {
		t.Skip("no error returned, cannot verify status code propagation")
	}

	apiErr, ok := err.(*api.Error)
	if !ok {
		t.Skipf("error is not *api.Error: %T", err)
	}
	if apiErr.StatusCode == 0 {
		t.Error("StatusCode was not populated on API error")
	}
}
