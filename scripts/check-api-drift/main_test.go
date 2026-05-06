package main

import "testing"

func TestParseOpenAPIYAMLExtractsEndpointsAndSchemaFields(t *testing.T) {
	data := []byte(`
openapi: 3.1.0
paths:
  /v2/projects/{project_id}/subscriptions/{subscription_id}/actions/extend:
    post:
      operationId: extend-subscription
components:
  schemas:
    Subscription:
      type: object
      properties:
        id:
          type: string
        current_period_ends_at:
          type: integer
`)
	endpoints, fields := parseOpenAPIYAML(data)
	if len(endpoints) != 1 {
		t.Fatalf("endpoints = %#v, want one endpoint", endpoints)
	}
	if got, want := endpointKey(endpoints[0]), "POST /projects/{}/subscriptions/{}/actions/extend"; got != want {
		t.Fatalf("endpoint = %q, want %q", got, want)
	}
	for _, field := range []string{"id", "current_period_ends_at"} {
		if _, ok := fields["Subscription"][field]; !ok {
			t.Fatalf("Subscription fields missing %q: %#v", field, fields["Subscription"])
		}
	}
}

func TestEndpointsFromGoFileFindsGenericPaginationAndAssertions(t *testing.T) {
	data := []byte(`
package fixture

import "github.com/andresdefi/rc/internal/api"

func run(client *api.Client) {
	path := "/projects/%s/integrations/webhooks"
	_, _ = api.PaginateAll[api.Webhook](client, path, nil)
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/subscriptions/sub_cmdtest/actions/extend")
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_cmdtest/charts/revenue")
}
`)
	got := map[string]bool{}
	for _, ep := range endpointsFromGoFile("fixture_test.go", data) {
		got[endpointKey(ep)] = true
	}
	for _, key := range []string{
		"GET /projects/{}/integrations/webhooks",
		"POST /projects/{}/subscriptions/{}/actions/extend",
		"GET /projects/{}/charts/{}",
	} {
		if !got[key] {
			t.Fatalf("missing endpoint %q in %#v", key, got)
		}
	}
}

func TestDocumentationKeyUsesResourceGroup(t *testing.T) {
	tests := map[string]string{
		"/projects":                   "projects",
		"/projects/{project_id}/apps": "apps",
		"/projects/{project_id}/integrations/webhooks":     "webhooks",
		"/projects/{project_id}/metrics/overview":          "charts",
		"/projects/{project_id}/virtual_currencies/{code}": "currencies",
	}
	for path, want := range tests {
		if got := documentationKey(path); got != want {
			t.Fatalf("documentationKey(%q) = %q, want %q", path, got, want)
		}
	}
}
