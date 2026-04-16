package cmd_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestFixtureAPIErrorSurfacesTypeMessageAndDocURL(t *testing.T) {
	result := cmdtest.Run(t, []string{
		"products", "create",
		"--store-id", "com.example.invalid",
		"--app-id", "app_cmdtest",
		"--type", "subscription",
	}, cmdtest.WithHandler(func(w http.ResponseWriter, r *http.Request) {
		writeFixtureJSON(w, http.StatusBadRequest, map[string]any{
			"object":  "error",
			"type":    "parameter_error",
			"message": "store identifier is invalid",
			"doc_url": "https://www.revenuecat.com/docs/api-v2",
		})
	}))

	cmdtest.AssertErrorContains(t, result, "parameter_error: store identifier is invalid")
	cmdtest.AssertErrorContains(t, result, "https://www.revenuecat.com/docs/api-v2")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/products")
}

func TestFixtureRetryableAPIErrorRetriesWithSameRequestBody(t *testing.T) {
	attempts := 0
	result := cmdtest.Run(t, []string{
		"products", "create",
		"--store-id", "com.example.retry",
		"--app-id", "app_cmdtest",
		"--type", "subscription",
		"--display-name", "Retry Product",
	}, cmdtest.WithHandler(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/projects/proj_cmdtest/products" {
			attempts++
			if attempts == 1 {
				backoffMs := 1
				writeFixtureJSON(w, http.StatusTooManyRequests, map[string]any{
					"object":     "error",
					"type":       "rate_limit_error",
					"message":    "rate limited",
					"retryable":  true,
					"backoff_ms": backoffMs,
				})
				return
			}
		}
		cmdtest.DefaultHandler(w, r)
	}))

	cmdtest.AssertSuccess(t, result)
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if got := countRequests(result, http.MethodPost, "/projects/proj_cmdtest/products"); got != 2 {
		t.Fatalf("POST /products requests = %d, want 2", got)
	}
	want := map[string]any{
		"store_identifier": "com.example.retry",
		"app_id":           "app_cmdtest",
		"type":             "subscription",
		"display_name":     "Retry Product",
	}
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/products", want)
	for _, req := range result.Requests {
		if req.Method == http.MethodPost && req.Path == "/projects/proj_cmdtest/products" && !strings.Contains(req.Body, "com.example.retry") {
			t.Fatalf("retry request body lost original payload: %#v", req)
		}
	}
}

func TestImportRestoresArchivedResourceState(t *testing.T) {
	file := writeGoldenFile(t, "archived-project.json", `{
  "version": "2",
  "apps": [
    {"object":"app","id":"app_source","name":"iOS App","type":"ios","project_id":"proj_source","created_at":1713072000000}
  ],
  "products": [
    {"object":"product","id":"prod_source","store_identifier":"com.example.archived","type":"subscription","state":"archived","display_name":"Archived Product","app_id":"app_source","created_at":1713072000000}
  ],
  "entitlements": [
    {"object":"entitlement","id":"entl_source","lookup_key":"archived","display_name":"Archived Access","state":"archived"}
  ],
  "offerings": [
    {"object":"offering","id":"ofrnge_source","lookup_key":"archived","display_name":"Archived Offering","state":"archived"}
  ]
}`)

	result := cmdtest.Run(t, []string{"import", "--file", file, "--app-map", "app_source=app_cmdtest"})

	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequestBody(t, result, http.MethodPost, "/projects/proj_cmdtest/products/prod_cmdtest/actions/archive", "")
	cmdtest.AssertRequestBody(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements/entl_cmdtest/actions/archive", "")
	cmdtest.AssertRequestBody(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/actions/archive", "")
}

func TestProjectImportReusesExistingResourcesWithoutDuplicateCreates(t *testing.T) {
	file := writeGoldenFile(t, "idempotent-project.json", `{
  "version": "2",
  "apps": [
    {"object":"app","id":"app_source","name":"iOS App","type":"ios","project_id":"proj_source","created_at":1713072000000}
  ],
  "products": [
    {"object":"product","id":"prod_source","store_identifier":"com.example.premium.monthly","type":"subscription","state":"active","display_name":"Premium Monthly","app_id":"app_source","created_at":1713072000000}
  ],
  "entitlements": [
    {"object":"entitlement","id":"entl_source","lookup_key":"premium","display_name":"Premium","state":"active","product_ids":["prod_source"]}
  ],
  "offerings": [
    {"object":"offering","id":"ofrnge_source","lookup_key":"default","display_name":"Default","is_current":false,"state":"active","packages":[
      {"object":"package","id":"pkge_source","lookup_key":"$rc_monthly","display_name":"Monthly","position":1,"products":[{"product_id":"prod_source","eligibility_criteria":"all"}]}
    ]}
  ]
}`)

	result := cmdtest.Run(t, []string{"import", "--file", file, "--app-map", "app_source=app_target"}, cmdtest.WithHandler(idempotentProjectImportHandler))

	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertNotRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/products")
	cmdtest.AssertNotRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements")
	cmdtest.AssertNotRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings")
	cmdtest.AssertNotRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_target/packages")
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements/entl_target", map[string]any{
		"display_name": "Premium",
	})
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements/entl_target/actions/attach_products", map[string]any{
		"product_ids": []any{"prod_target"},
	})
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_target", map[string]any{
		"display_name": "Default",
	})
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/packages/pkge_target", map[string]any{
		"display_name": "Monthly",
		"position":     float64(1),
	})
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/packages/pkge_target/actions/attach_products", map[string]any{
		"products": []any{
			map[string]any{
				"product_id":           "prod_target",
				"eligibility_criteria": "all",
			},
		},
	})
}

func idempotentProjectImportHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/projects/proj_cmdtest/apps":
		writeFixtureJSON(w, http.StatusOK, listFixture(map[string]any{
			"object": "app", "id": "app_target", "name": "iOS App", "type": "ios", "project_id": "proj_cmdtest", "created_at": 1713072000000,
		}))
	case r.Method == http.MethodGet && r.URL.Path == "/projects/proj_cmdtest/products":
		writeFixtureJSON(w, http.StatusOK, listFixture(map[string]any{
			"object": "product", "id": "prod_target", "store_identifier": "com.example.premium.monthly", "type": "subscription", "state": "active", "display_name": "Premium Monthly", "app_id": "app_target", "created_at": 1713072000000,
		}))
	case r.Method == http.MethodGet && r.URL.Path == "/projects/proj_cmdtest/entitlements":
		writeFixtureJSON(w, http.StatusOK, listFixture(map[string]any{
			"object": "entitlement", "id": "entl_target", "project_id": "proj_cmdtest", "lookup_key": "premium", "display_name": "Premium", "state": "active", "created_at": 1713072000000,
		}))
	case r.Method == http.MethodGet && r.URL.Path == "/projects/proj_cmdtest/offerings":
		writeFixtureJSON(w, http.StatusOK, listFixture(map[string]any{
			"object": "offering", "id": "ofrnge_target", "project_id": "proj_cmdtest", "lookup_key": "default", "display_name": "Default", "is_current": false, "state": "active", "created_at": 1713072000000,
		}))
	case r.Method == http.MethodGet && r.URL.Path == "/projects/proj_cmdtest/offerings/ofrnge_target/packages":
		position := 1
		writeFixtureJSON(w, http.StatusOK, listFixture(map[string]any{
			"object": "package", "id": "pkge_target", "lookup_key": "$rc_monthly", "display_name": "Monthly", "position": position, "created_at": 1713072000000,
		}))
	case r.Method == http.MethodPost && r.URL.Path == "/projects/proj_cmdtest/entitlements/entl_target":
		writeFixtureJSON(w, http.StatusOK, map[string]any{
			"object": "entitlement", "id": "entl_target", "project_id": "proj_cmdtest", "lookup_key": "premium", "display_name": "Premium", "state": "active", "created_at": 1713072000000,
		})
	case r.Method == http.MethodPost && r.URL.Path == "/projects/proj_cmdtest/offerings/ofrnge_target":
		writeFixtureJSON(w, http.StatusOK, map[string]any{
			"object": "offering", "id": "ofrnge_target", "project_id": "proj_cmdtest", "lookup_key": "default", "display_name": "Default", "is_current": false, "state": "active", "created_at": 1713072000000,
		})
	case r.Method == http.MethodPost && r.URL.Path == "/projects/proj_cmdtest/packages/pkge_target":
		position := 1
		writeFixtureJSON(w, http.StatusOK, map[string]any{
			"object": "package", "id": "pkge_target", "lookup_key": "$rc_monthly", "display_name": "Monthly", "position": position, "created_at": 1713072000000,
		})
	case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/actions/attach_products"):
		writeFixtureJSON(w, http.StatusOK, map[string]any{"object": "action_result", "id": "ok"})
	default:
		cmdtest.DefaultHandler(w, r)
	}
}

func countRequests(result cmdtest.Result, method, path string) int {
	count := 0
	for _, req := range result.Requests {
		if req.Method == method && req.Path == path {
			count++
		}
	}
	return count
}

func listFixture(items ...any) map[string]any {
	return map[string]any{
		"object":    "list",
		"items":     items,
		"next_page": nil,
		"url":       "/fixture",
	}
}

func writeFixtureJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
