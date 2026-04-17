package cmd_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestGoldenCreateRequestBodies(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
		want map[string]any
	}{
		{
			name: "projects create",
			args: []string{"projects", "create", "--name", "Golden Project"},
			path: "/projects",
			want: map[string]any{"name": "Golden Project"},
		},
		{
			name: "apps create",
			args: []string{"apps", "create", "--name", "iOS App", "--type", "app_store", "--bundle-id", "com.example.app"},
			path: "/projects/proj_cmdtest/apps",
			want: map[string]any{
				"name":      "iOS App",
				"type":      "app_store",
				"app_store": map[string]any{"bundle_id": "com.example.app"},
			},
		},
		{
			name: "products create",
			args: []string{"products", "create", "--store-id", "com.example.premium.yearly", "--app-id", "app_cmdtest", "--type", "subscription", "--display-name", "Premium Yearly"},
			path: "/projects/proj_cmdtest/products",
			want: map[string]any{
				"store_identifier": "com.example.premium.yearly",
				"app_id":           "app_cmdtest",
				"type":             "subscription",
				"display_name":     "Premium Yearly",
			},
		},
		{
			name: "entitlements create",
			args: []string{"entitlements", "create", "--lookup-key", "premium", "--display-name", "Premium Access"},
			path: "/projects/proj_cmdtest/entitlements",
			want: map[string]any{"lookup_key": "premium", "display_name": "Premium Access"},
		},
		{
			name: "offerings create",
			args: []string{"offerings", "create", "--lookup-key", "default", "--display-name", "Default Offering"},
			path: "/projects/proj_cmdtest/offerings",
			want: map[string]any{"lookup_key": "default", "display_name": "Default Offering"},
		},
		{
			name: "packages create",
			args: []string{"packages", "create", "--offering-id", "ofrnge_cmdtest", "--lookup-key", "$rc_annual", "--display-name", "Annual", "--position", "2"},
			path: "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/packages",
			want: map[string]any{"lookup_key": "$rc_annual", "display_name": "Annual", "position": float64(2)},
		},
		{
			name: "customers create",
			args: []string{"customers", "create", "--id", "cust_gold"},
			path: "/projects/proj_cmdtest/customers",
			want: map[string]any{"id": "cust_gold"},
		},
		{
			name: "webhooks create",
			args: []string{"webhooks", "create", "--name", "Events", "--url", "https://example.com/revenuecat"},
			path: "/projects/proj_cmdtest/integrations/webhooks",
			want: map[string]any{"name": "Events", "url": "https://example.com/revenuecat"},
		},
		{
			name: "paywalls create",
			args: []string{"paywalls", "create", "--offering-id", "ofrnge_cmdtest"},
			path: "/projects/proj_cmdtest/paywalls",
			want: map[string]any{"offering_id": "ofrnge_cmdtest"},
		},
		{
			name: "currencies create",
			args: []string{"currencies", "create", "--code", "COIN", "--name", "Coins"},
			path: "/projects/proj_cmdtest/virtual_currencies",
			want: map[string]any{"code": "COIN", "name": "Coins"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmdtest.Run(t, tt.args)
			cmdtest.AssertSuccess(t, result)
			cmdtest.AssertRequestJSON(t, result, http.MethodPost, tt.path, tt.want)
		})
	}
}

func TestGoldenUpdateRequestBodies(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
		want map[string]any
	}{
		{
			name: "apps update name",
			args: []string{"apps", "update", "app_cmdtest", "--name", "Renamed App"},
			path: "/projects/proj_cmdtest/apps/app_cmdtest",
			want: map[string]any{"name": "Renamed App"},
		},
		{
			name: "products update",
			args: []string{"products", "update", "prod_cmdtest", "--display-name", "Premium Monthly"},
			path: "/projects/proj_cmdtest/products/prod_cmdtest",
			want: map[string]any{"display_name": "Premium Monthly"},
		},
		{
			name: "entitlements update",
			args: []string{"entitlements", "update", "entl_cmdtest", "--display-name", "Premium Plus"},
			path: "/projects/proj_cmdtest/entitlements/entl_cmdtest",
			want: map[string]any{"display_name": "Premium Plus"},
		},
		{
			name: "offerings update",
			args: []string{"offerings", "update", "ofrnge_cmdtest", "--display-name", "Default Plus", "--is-current"},
			path: "/projects/proj_cmdtest/offerings/ofrnge_cmdtest",
			want: map[string]any{"display_name": "Default Plus", "is_current": true},
		},
		{
			name: "packages update",
			args: []string{"packages", "update", "pkge_cmdtest", "--display-name", "Annual Plus", "--position", "3"},
			path: "/projects/proj_cmdtest/packages/pkge_cmdtest",
			want: map[string]any{"display_name": "Annual Plus", "position": float64(3)},
		},
		{
			name: "webhooks update",
			args: []string{"webhooks", "update", "wh_cmdtest", "--name", "Events v2", "--url", "https://example.com/v2"},
			path: "/projects/proj_cmdtest/integrations/webhooks/wh_cmdtest",
			want: map[string]any{"name": "Events v2", "url": "https://example.com/v2"},
		},
		{
			name: "currencies update",
			args: []string{"currencies", "update", "COIN", "--name", "Coins Plus"},
			path: "/projects/proj_cmdtest/virtual_currencies/COIN",
			want: map[string]any{"name": "Coins Plus"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmdtest.Run(t, tt.args)
			cmdtest.AssertSuccess(t, result)
			cmdtest.AssertRequestJSON(t, result, http.MethodPost, tt.path, tt.want)
		})
	}
}

func TestGoldenAppStoreCredentialUpdateRequestBody(t *testing.T) {
	keyFile := writeGoldenFile(t, "SubscriptionKey_ABC123.p8", "-----BEGIN PRIVATE KEY-----\nabc123\n-----END PRIVATE KEY-----\n")
	connectKeyFile := writeGoldenFile(t, "AuthKey_ABC123.p8", "-----BEGIN PRIVATE KEY-----\nconnect123\n-----END PRIVATE KEY-----\n")

	result := cmdtest.Run(t, []string{
		"apps", "update", "app_cmdtest",
		"--shared-secret", "shared_secret",
		"--subscription-key-file", keyFile,
		"--subscription-key-id", "ABC123",
		"--subscription-key-issuer", "issuer-123",
		"--app-store-connect-api-key-file", connectKeyFile,
		"--app-store-connect-api-key-id", "CONNECT123",
		"--app-store-connect-api-key-issuer", "connect-issuer-123",
		"--app-store-connect-vendor-number", "12345678",
	})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/apps/app_cmdtest", map[string]any{
		"app_store": map[string]any{
			"shared_secret":                    "shared_secret",
			"subscription_private_key":         "-----BEGIN PRIVATE KEY-----\nabc123\n-----END PRIVATE KEY-----\n",
			"subscription_key_id":              "ABC123",
			"subscription_key_issuer":          "issuer-123",
			"app_store_connect_api_key":        "-----BEGIN PRIVATE KEY-----\nconnect123\n-----END PRIVATE KEY-----\n",
			"app_store_connect_api_key_id":     "CONNECT123",
			"app_store_connect_api_key_issuer": "connect-issuer-123",
			"app_store_connect_vendor_number":  "12345678",
		},
	})
}

func TestGoldenArchiveActionRequestBodies(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
	}{
		{"products archive", []string{"products", "archive", "prod_cmdtest"}, "/projects/proj_cmdtest/products/prod_cmdtest/actions/archive"},
		{"products unarchive", []string{"products", "unarchive", "prod_cmdtest"}, "/projects/proj_cmdtest/products/prod_cmdtest/actions/unarchive"},
		{"entitlements archive", []string{"entitlements", "archive", "entl_cmdtest"}, "/projects/proj_cmdtest/entitlements/entl_cmdtest/actions/archive"},
		{"entitlements unarchive", []string{"entitlements", "unarchive", "entl_cmdtest"}, "/projects/proj_cmdtest/entitlements/entl_cmdtest/actions/unarchive"},
		{"offerings archive", []string{"offerings", "archive", "ofrnge_cmdtest"}, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/actions/archive"},
		{"offerings unarchive", []string{"offerings", "unarchive", "ofrnge_cmdtest"}, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/actions/unarchive"},
		{"currencies archive", []string{"currencies", "archive", "COIN"}, "/projects/proj_cmdtest/virtual_currencies/COIN/actions/archive"},
		{"currencies unarchive", []string{"currencies", "unarchive", "COIN"}, "/projects/proj_cmdtest/virtual_currencies/COIN/actions/unarchive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmdtest.Run(t, tt.args)
			cmdtest.AssertSuccess(t, result)
			cmdtest.AssertRequestBody(t, result, http.MethodPost, tt.path, "")
		})
	}
}

func TestGoldenOfferingUnarchiveReferencedEntitiesRequestBody(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "unarchive", "ofrnge_cmdtest", "--unarchive-referenced-entities"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/actions/unarchive", map[string]any{
		"unarchive_referenced_entities": true,
	})
}

func TestGoldenProductPushToStoreSubscriptionRequestBody(t *testing.T) {
	result := cmdtest.Run(t, []string{
		"products", "push-to-store", "prod_cmdtest",
		"--subscription-duration", "ONE_MONTH",
		"--subscription-group-name", "Premium Subscriptions",
		"--subscription-group-id", "sub_group_123",
	})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/products/prod_cmdtest/create_in_store", map[string]any{
		"store_information": map[string]any{
			"duration":                "ONE_MONTH",
			"subscription_group_name": "Premium Subscriptions",
			"subscription_group_id":   "sub_group_123",
		},
	})
}

func TestGoldenAttachDetachRequestBodies(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
		want map[string]any
	}{
		{
			name: "entitlements attach",
			args: []string{"entitlements", "attach", "--entitlement-id", "entl_cmdtest", "--product-id", "prod_monthly,prod_yearly"},
			path: "/projects/proj_cmdtest/entitlements/entl_cmdtest/actions/attach_products",
			want: map[string]any{"product_ids": []any{"prod_monthly", "prod_yearly"}},
		},
		{
			name: "entitlements detach",
			args: []string{"entitlements", "detach", "--entitlement-id", "entl_cmdtest", "--product-id", "prod_monthly,prod_yearly"},
			path: "/projects/proj_cmdtest/entitlements/entl_cmdtest/actions/detach_products",
			want: map[string]any{"product_ids": []any{"prod_monthly", "prod_yearly"}},
		},
		{
			name: "packages attach",
			args: []string{"packages", "attach", "--package-id", "pkge_cmdtest", "--product-id", "prod_monthly", "--eligibility", "google_sdk_ge_6"},
			path: "/projects/proj_cmdtest/packages/pkge_cmdtest/actions/attach_products",
			want: map[string]any{
				"products": []any{
					map[string]any{
						"product_id":           "prod_monthly",
						"eligibility_criteria": "google_sdk_ge_6",
					},
				},
			},
		},
		{
			name: "packages detach",
			args: []string{"packages", "detach", "--package-id", "pkge_cmdtest", "--product-id", "prod_monthly,prod_yearly"},
			path: "/projects/proj_cmdtest/packages/pkge_cmdtest/actions/detach_products",
			want: map[string]any{"product_ids": []any{"prod_monthly", "prod_yearly"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmdtest.Run(t, tt.args)
			cmdtest.AssertSuccess(t, result)
			cmdtest.AssertRequestJSON(t, result, http.MethodPost, tt.path, tt.want)
		})
	}
}

func TestGoldenImportRequestBodies(t *testing.T) {
	productFile := writeGoldenFile(t, "products.json", `[
  {"store_identifier":"com.example.premium.imported","app_id":"app_cmdtest","type":"subscription","display_name":"Imported Premium"}
]`)
	entitlementFile := writeGoldenFile(t, "entitlements.json", `[
  {"lookup_key":"imported","display_name":"Imported Access"}
]`)
	projectFile := writeGoldenFile(t, "project.json", `{
  "version": "2",
  "apps": [
    {"object":"app","id":"app_source","name":"iOS App","type":"ios","project_id":"proj_source","created_at":1713072000000}
  ],
  "products": [
    {"object":"product","id":"prod_source","store_identifier":"com.example.premium.imported","type":"subscription","state":"active","display_name":"Imported Premium","app_id":"app_source","created_at":1713072000000}
  ],
  "entitlements": [
    {"object":"entitlement","id":"entl_source","lookup_key":"imported","display_name":"Imported Access","state":"active","product_ids":["prod_source"]}
  ],
  "offerings": [
    {"object":"offering","id":"ofrnge_source","lookup_key":"imported","display_name":"Imported Offering","is_current":true,"state":"active","metadata":{"tier":"gold"},"packages":[
      {"object":"package","id":"pkge_source","lookup_key":"$rc_annual","display_name":"Annual","position":2,"products":[{"product_id":"prod_source","eligibility_criteria":"all"}]}
    ]}
  ]
}`)

	tests := []struct {
		name   string
		args   []string
		assert func(*testing.T, cmdtest.Result)
	}{
		{
			name: "products import",
			args: []string{"products", "import", "--file", productFile},
			assert: func(t *testing.T, result cmdtest.Result) {
				t.Helper()
				cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/products", map[string]any{
					"store_identifier": "com.example.premium.imported",
					"app_id":           "app_cmdtest",
					"type":             "subscription",
					"display_name":     "Imported Premium",
				})
			},
		},
		{
			name: "entitlements import",
			args: []string{"entitlements", "import", "--file", entitlementFile},
			assert: func(t *testing.T, result cmdtest.Result) {
				t.Helper()
				cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements", map[string]any{
					"lookup_key":   "imported",
					"display_name": "Imported Access",
				})
			},
		},
		{
			name: "project import",
			args: []string{"import", "--file", projectFile, "--app-map", "app_source=app_cmdtest"},
			assert: func(t *testing.T, result cmdtest.Result) {
				t.Helper()
				cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/products", map[string]any{
					"store_identifier": "com.example.premium.imported",
					"app_id":           "app_cmdtest",
					"type":             "subscription",
					"display_name":     "Imported Premium",
				})
				cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements", map[string]any{
					"lookup_key":   "imported",
					"display_name": "Imported Access",
				})
				cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements/entl_cmdtest/actions/attach_products", map[string]any{
					"product_ids": []any{"prod_cmdtest"},
				})
				cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings", map[string]any{
					"lookup_key":   "imported",
					"display_name": "Imported Offering",
					"metadata":     map[string]any{"tier": "gold"},
				})
				cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/packages", map[string]any{
					"lookup_key":   "$rc_annual",
					"display_name": "Annual",
					"position":     float64(2),
				})
				cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/packages/pkge_cmdtest/actions/attach_products", map[string]any{
					"products": []any{
						map[string]any{
							"product_id":           "prod_cmdtest",
							"eligibility_criteria": "all",
						},
					},
				})
				cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest", map[string]any{"is_current": true})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmdtest.Run(t, tt.args)
			cmdtest.AssertSuccess(t, result)
			tt.assert(t, result)
		})
	}
}

func writeGoldenFile(t *testing.T, name, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write fixture %s: %v", name, err)
	}
	return path
}
