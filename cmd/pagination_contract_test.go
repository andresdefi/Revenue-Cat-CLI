package cmd_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestListCommandsAllFollowNextPagePath(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
	}{
		{name: "projects", args: []string{"projects", "list", "--all", "--output", "json"}, path: "/projects"},
		{name: "apps", args: []string{"apps", "list", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/apps"},
		{name: "products", args: []string{"products", "list", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/products"},
		{name: "entitlements", args: []string{"entitlements", "list", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/entitlements"},
		{name: "entitlement products", args: []string{"entitlements", "products", "entl_cmdtest", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/entitlements/entl_cmdtest/products"},
		{name: "offerings", args: []string{"offerings", "list", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/offerings"},
		{name: "packages", args: []string{"packages", "list", "--offering-id", "ofrnge_cmdtest", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/packages"},
		{name: "package products", args: []string{"packages", "products", "pkge_cmdtest", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/packages/pkge_cmdtest/products"},
		{name: "customers", args: []string{"customers", "list", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/customers"},
		{name: "customer subscriptions", args: []string{"customers", "subscriptions", "cust_cmdtest", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/customers/cust_cmdtest/subscriptions"},
		{name: "customer purchases", args: []string{"customers", "purchases", "cust_cmdtest", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/customers/cust_cmdtest/purchases"},
		{name: "subscriptions", args: []string{"subscriptions", "list", "--store-subscription-id", "store_sub_cmdtest", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/subscriptions"},
		{name: "subscription transactions", args: []string{"subscriptions", "transactions", "sub_cmdtest", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/subscriptions/sub_cmdtest/transactions"},
		{name: "purchases", args: []string{"purchases", "list", "--store-purchase-id", "store_purchase_cmdtest", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/purchases"},
		{name: "webhooks", args: []string{"webhooks", "list", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/integrations/webhooks"},
		{name: "paywalls", args: []string{"paywalls", "list", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/paywalls"},
		{name: "audit logs", args: []string{"audit-logs", "list", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/audit_logs"},
		{name: "collaborators", args: []string{"collaborators", "list", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/collaborators"},
		{name: "currencies", args: []string{"currencies", "list", "--all", "--output", "json"}, path: "/projects/proj_cmdtest/virtual_currencies"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmdtest.Run(t, tt.args, cmdtest.WithHandler(paginatedHandler(tt.path)))
			cmdtest.AssertSuccess(t, result)
			cmdtest.AssertRequested(t, result, http.MethodGet, tt.path)
			cmdtest.AssertRequestedWithQuery(t, result, http.MethodGet, tt.path, "starting_after", "cursor_1")
		})
	}
}

func paginatedHandler(expectedPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != expectedPath {
			writeList(w, nil)
			return
		}
		if r.URL.Query().Get("starting_after") == "" {
			writeList(w, stringPtr(expectedPath+"?starting_after=cursor_1"))
			return
		}
		writeList(w, nil)
	}
}

func writeList(w http.ResponseWriter, nextPage *string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"object":    "list",
		"items":     []any{map[string]any{"object": "fixture", "id": "fixture_id", "name": "Fixture"}},
		"next_page": nextPage,
		"url":       "/fixture",
	})
}

func stringPtr(s string) *string {
	return &s
}
