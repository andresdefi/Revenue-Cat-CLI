//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/andresdefi/rc/internal/api"
)

// These tests run against the real RevenueCat API.
// They are gated behind the "integration" build tag and require:
//
//	RC_INTEGRATION_KEY - a valid RevenueCat API v2 secret key (sk_...)
//
// Optional:
//
//	RC_INTEGRATION_PROJECT_ID - project used for project-scoped read tests
//
// Run with: go test -tags integration ./internal/integration/...
// Or:       make test-integration

func integrationClient(t *testing.T) *api.Client {
	t.Helper()
	token := os.Getenv("RC_INTEGRATION_KEY")
	if token == "" {
		t.Skip("RC_INTEGRATION_KEY is required for integration tests")
	}
	return api.NewClientWithToken(token)
}

func TestProjectsListIntegration(t *testing.T) {
	client := integrationClient(t)
	data, err := client.Get("/projects", nil)
	if err != nil {
		t.Fatalf("projects list: %v", err)
	}
	var resp api.ListResponse[api.Project]
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("unmarshal projects: %v", err)
	}
	if resp.Object != "list" {
		t.Fatalf("object = %q, want list", resp.Object)
	}
}

func TestProjectAppsListIntegration(t *testing.T) {
	projectID := os.Getenv("RC_INTEGRATION_PROJECT_ID")
	if projectID == "" {
		t.Skip("RC_INTEGRATION_PROJECT_ID is required for project-scoped integration tests")
	}
	client := integrationClient(t)
	items, err := api.PaginateAll[api.App](client, "/projects/"+projectID+"/apps", nil)
	if err != nil {
		t.Fatalf("project apps list: %v", err)
	}
	_ = items
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

func TestErrorStatusCodePropagation(t *testing.T) {
	client := api.NewClientWithToken("sk_test_status_code_check")

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
