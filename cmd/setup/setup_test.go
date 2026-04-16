package setup_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestSetupProductReusesExistingPath(t *testing.T) {
	result := cmdtest.Run(t, []string{
		"setup", "product",
		"--app-id", "app_cmdtest",
		"--store-id", "com.example.premium.monthly",
		"--output", "json",
	})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"object": "product_setup"`)
	cmdtest.AssertOutputContains(t, result, `"status": "ok"`)
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_cmdtest/products")
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_cmdtest/entitlements")
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_cmdtest/offerings")
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/packages")

	for _, req := range result.Requests {
		if req.Method == http.MethodPost {
			t.Fatalf("expected existing setup path to avoid mutations, got POST %s", req.Path)
		}
	}
}

func TestSetupProductCreatesMissingPath(t *testing.T) {
	result := cmdtest.Run(t,
		[]string{
			"setup", "product",
			"--app-id", "app_new",
			"--store-id", "com.example.pro.monthly",
			"--display-name", "Pro Monthly",
			"--entitlement-key", "pro",
			"--offering-key", "pro",
			"--package-key", "$rc_monthly",
			"--make-current",
			"--output", "json",
		},
		cmdtest.WithHandler(setupCreateHandler),
	)
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"status": "changed"`)
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/products")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_new/packages")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements/entl_new/actions/attach_products")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/packages/pkge_new/actions/attach_products")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_new")
}

func TestSetupProductDryRunDoesNotMutate(t *testing.T) {
	result := cmdtest.Run(t, []string{
		"--dry-run",
		"setup", "product",
		"--app-id", "app_missing",
		"--store-id", "com.example.dryrun",
		"--output", "json",
	}, cmdtest.WithHandler(setupEmptyHandler))
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"dry_run": true`)
	cmdtest.AssertOutputContains(t, result, `"status": "planned"`)

	for _, req := range result.Requests {
		if req.Method == http.MethodPost || req.Method == http.MethodDelete {
			t.Fatalf("dry-run should not mutate, got %s %s", req.Method, req.Path)
		}
	}
}

func setupEmptyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeSetupJSON(w, map[string]any{"object": "error", "type": "unexpected", "message": "unexpected mutation"})
		return
	}
	writeSetupJSON(w, setupList())
}

func setupCreateHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet:
		writeSetupJSON(w, setupList())
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/products"):
		writeSetupJSON(w, map[string]any{"object": "product", "id": "prod_new", "store_identifier": "com.example.pro.monthly", "type": "subscription", "state": "active", "display_name": "Pro Monthly", "app_id": "app_new", "created_at": 1713072000000})
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/entitlements"):
		writeSetupJSON(w, map[string]any{"object": "entitlement", "id": "entl_new", "project_id": cmdtest.TestProjectID, "lookup_key": "pro", "display_name": "pro", "state": "active", "created_at": 1713072000000})
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/offerings"):
		writeSetupJSON(w, map[string]any{"object": "offering", "id": "ofrnge_new", "project_id": cmdtest.TestProjectID, "lookup_key": "pro", "display_name": "pro", "is_current": false, "state": "active", "created_at": 1713072000000})
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/packages"):
		writeSetupJSON(w, map[string]any{"object": "package", "id": "pkge_new", "lookup_key": "$rc_monthly", "display_name": "Monthly", "created_at": 1713072000000})
	case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/actions/attach_products"):
		writeSetupJSON(w, map[string]any{"object": "ok"})
	case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/offerings/ofrnge_new"):
		writeSetupJSON(w, map[string]any{"object": "offering", "id": "ofrnge_new", "project_id": cmdtest.TestProjectID, "lookup_key": "pro", "display_name": "pro", "is_current": true, "state": "active", "created_at": 1713072000000})
	default:
		writeSetupJSON(w, map[string]any{"object": "error", "type": "not_found", "message": "not found"})
	}
}

func setupList(items ...any) map[string]any {
	return map[string]any{"object": "list", "items": items, "next_page": nil, "url": "/fixture"}
}

func writeSetupJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
