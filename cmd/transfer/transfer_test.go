package transfer_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	transfercmd "github.com/andresdefi/rc/cmd/transfer"
	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestNewExportCmd_NotNil(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := transfercmd.NewExportCmd(&projectID, &outputFormat)
	if cmd == nil {
		t.Fatal("NewExportCmd() returned nil")
	}
}

func TestNewExportCmd_Use(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := transfercmd.NewExportCmd(&projectID, &outputFormat)
	if cmd.Use != "export" {
		t.Errorf("Use = %q, want %q", cmd.Use, "export")
	}
}

func TestNewExportCmd_Short(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := transfercmd.NewExportCmd(&projectID, &outputFormat)
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestNewExportCmd_HasFileFlag(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := transfercmd.NewExportCmd(&projectID, &outputFormat)
	f := cmd.Flags().Lookup("file")
	if f == nil {
		t.Fatal("export command missing --file flag")
	}
	if f.DefValue != "" {
		t.Errorf("--file default = %q, want empty string", f.DefValue)
	}
}

func TestNewImportCmd_NotNil(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := transfercmd.NewImportCmd(&projectID, &outputFormat)
	if cmd == nil {
		t.Fatal("NewImportCmd() returned nil")
	}
}

func TestNewImportCmd_Use(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := transfercmd.NewImportCmd(&projectID, &outputFormat)
	if cmd.Use != "import" {
		t.Errorf("Use = %q, want %q", cmd.Use, "import")
	}
}

func TestNewImportCmd_Short(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := transfercmd.NewImportCmd(&projectID, &outputFormat)
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestNewImportCmd_HasFileFlag(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := transfercmd.NewImportCmd(&projectID, &outputFormat)
	f := cmd.Flags().Lookup("file")
	if f == nil {
		t.Fatal("import command missing --file flag")
	}
	if f.DefValue != "" {
		t.Errorf("--file default = %q, want empty string", f.DefValue)
	}
}

func TestNewImportCmd_HasAppMapFlag(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := transfercmd.NewImportCmd(&projectID, &outputFormat)
	f := cmd.Flags().Lookup("app-map")
	if f == nil {
		t.Fatal("import command missing --app-map flag")
	}
}

func TestNewExportCmd_Long(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := transfercmd.NewExportCmd(&projectID, &outputFormat)
	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestExportProjectConfig_IncludesRelationships(t *testing.T) {
	file := filepath.Join(t.TempDir(), "project-config.json")

	result := cmdtest.Run(t, []string{"export", "--file", file})
	cmdtest.AssertSuccess(t, result)

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var config transfercmd.ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if config.Version != "2" {
		t.Errorf("Version = %q, want 2", config.Version)
	}
	if len(config.Apps) != 1 {
		t.Fatalf("Apps count = %d, want 1", len(config.Apps))
	}
	if len(config.Products) != 1 {
		t.Fatalf("Products count = %d, want 1", len(config.Products))
	}
	if len(config.Entitlements) != 1 {
		t.Fatalf("Entitlements count = %d, want 1", len(config.Entitlements))
	}
	if got := config.Entitlements[0].ProductIDs; len(got) != 1 || got[0] != "prod_cmdtest" {
		t.Fatalf("Entitlement ProductIDs = %#v, want [prod_cmdtest]", got)
	}
	if len(config.Offerings) != 1 {
		t.Fatalf("Offerings count = %d, want 1", len(config.Offerings))
	}
	if len(config.Offerings[0].Packages) != 1 {
		t.Fatalf("Packages count = %d, want 1", len(config.Offerings[0].Packages))
	}
	if got := config.Offerings[0].Packages[0].Products; len(got) != 1 || got[0].ProductID != "prod_cmdtest" {
		t.Fatalf("Package products = %#v, want product prod_cmdtest", got)
	}
}

func TestExportProjectConfig_IncludesArchiveStateAndMetadata(t *testing.T) {
	file := filepath.Join(t.TempDir(), "project-config.json")

	result := cmdtest.Run(t, []string{"export", "--file", file}, cmdtest.WithHandler(archivedMigrationFixtureHandler))
	cmdtest.AssertSuccess(t, result)

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var config transfercmd.ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := config.Products[0].State; got != "archived" {
		t.Fatalf("Product state = %q, want archived", got)
	}
	if got := config.Entitlements[0].State; got != "archived" {
		t.Fatalf("Entitlement state = %q, want archived", got)
	}
	if got := config.Offerings[0].State; got != "archived" {
		t.Fatalf("Offering state = %q, want archived", got)
	}
	if got := config.Offerings[0].Metadata["tier"]; got != "gold" {
		t.Fatalf("Offering metadata tier = %#v, want gold", got)
	}
}

func TestImportProjectConfig_AttachesMappedProducts(t *testing.T) {
	file := filepath.Join(t.TempDir(), "project-config.json")
	data := []byte(`{
  "version": "2",
  "apps": [
    {"object": "app", "id": "app_source", "name": "iOS App", "type": "ios", "project_id": "proj_source", "created_at": 1713072000000}
  ],
  "products": [
    {"object": "product", "id": "prod_source", "store_identifier": "com.example.premium.monthly", "type": "subscription", "state": "active", "display_name": "Premium Monthly", "app_id": "app_source", "created_at": 1713072000000}
  ],
  "entitlements": [
    {"object": "entitlement", "id": "entl_source", "lookup_key": "premium", "display_name": "Premium", "state": "active", "product_ids": ["prod_source"]}
  ],
  "offerings": [
    {"object": "offering", "id": "ofrnge_source", "lookup_key": "default", "display_name": "Default", "is_current": true, "state": "active", "packages": [
      {"object": "package", "id": "pkge_source", "lookup_key": "$rc_monthly", "display_name": "Monthly", "products": [
        {"product_id": "prod_source", "eligibility_criteria": "all"}
      ]}
    ]}
  ]
}`)
	if err := os.WriteFile(file, data, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	result := cmdtest.Run(t, []string{"import", "--file", file, "--app-map", "app_source=app_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements/entl_cmdtest/actions/attach_products")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/packages/pkge_cmdtest/actions/attach_products")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest")
}

func TestImportProjectConfig_IsIdempotentForExistingResources(t *testing.T) {
	file := writeTransferFixture(t, `{
  "version": "2",
  "apps": [
    {"object": "app", "id": "app_source", "name": "iOS App", "type": "ios", "project_id": "proj_source", "created_at": 1713072000000}
  ],
  "products": [
    {"object": "product", "id": "prod_source", "store_identifier": "com.example.premium.monthly", "type": "subscription", "state": "active", "display_name": "Premium Monthly", "app_id": "app_source", "created_at": 1713072000000}
  ],
  "entitlements": [
    {"object": "entitlement", "id": "entl_source", "lookup_key": "premium", "display_name": "Premium", "state": "active", "product_ids": ["prod_source"]}
  ],
  "offerings": [
    {"object": "offering", "id": "ofrnge_source", "lookup_key": "default", "display_name": "Default Offering", "is_current": true, "state": "active", "packages": [
      {"object": "package", "id": "pkge_source", "lookup_key": "$rc_monthly", "display_name": "Monthly", "products": [
        {"product_id": "prod_source", "eligibility_criteria": "all"}
      ]}
    ]}
  ]
}`)

	result := cmdtest.Run(t, []string{"import", "--file", file, "--app-map", "app_source=app_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertNotRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/products")
	cmdtest.AssertNotRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements")
	cmdtest.AssertNotRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings")
	cmdtest.AssertNotRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/packages")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements/entl_cmdtest/actions/attach_products")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/packages/pkge_cmdtest/actions/attach_products")
}

func TestImportProjectConfig_UpdatesExistingOfferingAndPackageBodies(t *testing.T) {
	file := writeTransferFixture(t, `{
  "version": "2",
  "apps": [
    {"object": "app", "id": "app_source", "name": "iOS App", "type": "ios", "project_id": "proj_source", "created_at": 1713072000000}
  ],
  "products": [
    {"object": "product", "id": "prod_source", "store_identifier": "com.example.premium.monthly", "type": "subscription", "state": "active", "display_name": "Premium Monthly", "app_id": "app_source", "created_at": 1713072000000}
  ],
  "entitlements": [],
  "offerings": [
    {"object": "offering", "id": "ofrnge_source", "lookup_key": "default", "display_name": "Default Plus", "is_current": true, "state": "active", "metadata":{"tier":"gold"},"packages": [
      {"object": "package", "id": "pkge_source", "lookup_key": "$rc_monthly", "display_name": "Monthly Plus", "position": 2, "products": [
        {"product_id": "prod_source", "eligibility_criteria": "google_sdk_ge_6"}
      ]}
    ]}
  ]
}`)

	result := cmdtest.Run(t, []string{"import", "--file", file, "--app-map", "app_source=app_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertNotRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings")
	cmdtest.AssertNotRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/packages")
	assertRequestJSONAmong(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest", map[string]any{
		"display_name": "Default Plus",
		"metadata":     map[string]any{"tier": "gold"},
	})
	assertRequestJSONAmong(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest", map[string]any{
		"is_current": true,
	})
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/packages/pkge_cmdtest", map[string]any{
		"display_name": "Monthly Plus",
		"position":     float64(2),
	})
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/packages/pkge_cmdtest/actions/attach_products", map[string]any{
		"products": []any{
			map[string]any{
				"product_id":           "prod_cmdtest",
				"eligibility_criteria": "google_sdk_ge_6",
			},
		},
	})
}

func TestImportProjectConfig_ArchivesImportedEntities(t *testing.T) {
	file := writeTransferFixture(t, `{
  "version": "2",
  "apps": [
    {"object": "app", "id": "app_source", "name": "iOS App", "type": "ios", "project_id": "proj_source", "created_at": 1713072000000}
  ],
  "products": [
    {"object": "product", "id": "prod_source", "store_identifier": "com.example.premium.monthly", "type": "subscription", "state": "archived", "display_name": "Premium Monthly", "app_id": "app_source", "created_at": 1713072000000}
  ],
  "entitlements": [
    {"object": "entitlement", "id": "entl_source", "lookup_key": "premium", "display_name": "Premium", "state": "archived", "product_ids": ["prod_source"]}
  ],
  "offerings": [
    {"object": "offering", "id": "ofrnge_source", "lookup_key": "default", "display_name": "Default Offering", "is_current": false, "state": "archived"}
  ]
}`)

	result := cmdtest.Run(t, []string{"import", "--file", file, "--app-map", "app_source=app_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/actions/archive")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements/entl_cmdtest/actions/archive")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/products/prod_cmdtest/actions/archive")
}

func TestImportProjectConfig_WarnsWhenAttachmentCallsFail(t *testing.T) {
	file := writeTransferFixture(t, `{
  "version": "2",
  "apps": [
    {"object": "app", "id": "app_source", "name": "iOS App", "type": "ios", "project_id": "proj_source", "created_at": 1713072000000}
  ],
  "products": [
    {"object": "product", "id": "prod_source", "store_identifier": "com.example.premium.monthly", "type": "subscription", "state": "active", "display_name": "Premium Monthly", "app_id": "app_source", "created_at": 1713072000000}
  ],
  "entitlements": [
    {"object": "entitlement", "id": "entl_source", "lookup_key": "premium", "display_name": "Premium", "state": "active", "product_ids": ["prod_source"]}
  ],
  "offerings": [
    {"object": "offering", "id": "ofrnge_source", "lookup_key": "default", "display_name": "Default Offering", "is_current": false, "state": "active", "packages": [
      {"object": "package", "id": "pkge_source", "lookup_key": "$rc_monthly", "display_name": "Monthly", "products": [
        {"product_id": "prod_source", "eligibility_criteria": "all"}
      ]}
    ]}
  ]
}`)

	result := cmdtest.Run(t, []string{"import", "--file", file, "--app-map", "app_source=app_cmdtest"}, cmdtest.WithHandler(failAttachmentHandler))
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Failed to attach products to entitlement premium")
	cmdtest.AssertOutputContains(t, result, "Failed to attach products to package $rc_monthly")
}

func TestImportProjectConfig_WarnsWhenArchiveCallsFail(t *testing.T) {
	file := writeTransferFixture(t, `{
  "version": "2",
  "apps": [
    {"object": "app", "id": "app_source", "name": "iOS App", "type": "ios", "project_id": "proj_source", "created_at": 1713072000000}
  ],
  "products": [
    {"object": "product", "id": "prod_source", "store_identifier": "com.example.premium.monthly", "type": "subscription", "state": "archived", "display_name": "Premium Monthly", "app_id": "app_source", "created_at": 1713072000000}
  ],
  "entitlements": [
    {"object": "entitlement", "id": "entl_source", "lookup_key": "premium", "display_name": "Premium", "state": "archived", "product_ids": ["prod_source"]}
  ],
  "offerings": [
    {"object": "offering", "id": "ofrnge_source", "lookup_key": "default", "display_name": "Default Offering", "is_current": false, "state": "archived"}
  ]
}`)

	result := cmdtest.Run(t, []string{"import", "--file", file, "--app-map", "app_source=app_cmdtest"}, cmdtest.WithHandler(failArchiveHandler))
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Failed to archive offering default")
	cmdtest.AssertOutputContains(t, result, "Failed to archive entitlement premium")
	cmdtest.AssertOutputContains(t, result, "Failed to archive product com.example.premium.monthly")
}

func TestImportProjectConfig_ContinuesAfterProductCreateFailure(t *testing.T) {
	file := writeTransferFixture(t, `{
  "version": "2",
  "apps": [
    {"object": "app", "id": "app_source", "name": "iOS App", "type": "ios", "project_id": "proj_source", "created_at": 1713072000000}
  ],
  "products": [
    {"object": "product", "id": "prod_bad_source", "store_identifier": "com.example.bad", "type": "subscription", "state": "active", "display_name": "Bad Product", "app_id": "app_source", "created_at": 1713072000000},
    {"object": "product", "id": "prod_good_source", "store_identifier": "com.example.good", "type": "subscription", "state": "active", "display_name": "Good Product", "app_id": "app_source", "created_at": 1713072000000}
  ],
  "entitlements": [
    {"object": "entitlement", "id": "entl_source", "lookup_key": "premium", "display_name": "Premium", "state": "active", "product_ids": ["prod_bad_source", "prod_good_source"]}
  ],
  "offerings": []
}`)

	result := cmdtest.Run(t, []string{"import", "--file", file, "--app-map", "app_source=app_cmdtest"}, cmdtest.WithHandler(newPartialImportFailureHandler()))
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Failed to create product com.example.bad")
	cmdtest.AssertOutputContains(t, result, "Skipping entitlement product attachment premium -> prod_bad_source")
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/entitlements/entl_target/actions/attach_products", map[string]any{
		"product_ids": []any{"prod_good_target"},
	})
}

func TestNewImportCmd_Long(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := transfercmd.NewImportCmd(&projectID, &outputFormat)
	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestMigrateProjectRequiresDryRun(t *testing.T) {
	result := cmdtest.Run(t, []string{"migrate", "project", "--source-project", "proj_source"})
	cmdtest.AssertErrorContains(t, result, "requires --dry-run")
}

func TestMigrateProjectDryRunPlan(t *testing.T) {
	result := cmdtest.Run(t, []string{"migrate", "project", "--source-project", "proj_source", "--target-project", "proj_target", "--dry-run", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"object": "project_migration_plan"`)
	cmdtest.AssertOutputContains(t, result, `"source_project_id": "proj_source"`)
	cmdtest.AssertOutputContains(t, result, `"target_project_id": "proj_target"`)
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_source/products")
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_target/products")

	for _, req := range result.Requests {
		if req.Method == http.MethodPost || req.Method == http.MethodDelete {
			t.Fatalf("migrate project --dry-run should not mutate, got %s %s", req.Method, req.Path)
		}
	}
}

func TestMigrateProjectDryRunReportsArchiveActions(t *testing.T) {
	result := cmdtest.Run(t, []string{
		"migrate", "project",
		"--source-project", "proj_source",
		"--target-project", "proj_target",
		"--dry-run",
		"--output", "json",
	}, cmdtest.WithHandler(archivedMigrationFixtureHandler))
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"archive": 3`)
	cmdtest.AssertOutputContains(t, result, `"area": "products"`)
	cmdtest.AssertOutputContains(t, result, `"area": "entitlements"`)
	cmdtest.AssertOutputContains(t, result, `"area": "offerings"`)

	for _, req := range result.Requests {
		if req.Method == http.MethodPost || req.Method == http.MethodDelete {
			t.Fatalf("migrate project --dry-run should not mutate, got %s %s", req.Method, req.Path)
		}
	}
}

func writeTransferFixture(t *testing.T, data string) string {
	t.Helper()
	file := filepath.Join(t.TempDir(), "project-config.json")
	if err := os.WriteFile(file, []byte(data), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return file
}

func assertRequestJSONAmong(t *testing.T, result cmdtest.Result, method, requestPath string, want map[string]any) {
	t.Helper()
	wantRaw, _ := json.Marshal(want)
	for _, req := range result.Requests {
		if req.Method != method || req.Path != requestPath {
			continue
		}
		var got map[string]any
		if err := json.Unmarshal([]byte(req.Body), &got); err != nil {
			t.Fatalf("request body for %s %s is not JSON: %v\nbody: %s", method, requestPath, err, req.Body)
		}
		gotRaw, _ := json.Marshal(got)
		if string(gotRaw) == string(wantRaw) {
			return
		}
	}
	t.Fatalf("missing request %s %s with body %s; got %#v", method, requestPath, wantRaw, result.Requests)
}

func failAttachmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/actions/attach_products") {
		writeTransferJSON(w, http.StatusInternalServerError, transferError("server_error", "temporary attachment failure"))
		return
	}
	cmdtest.DefaultHandler(w, r)
}

func failArchiveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/actions/archive") {
		writeTransferJSON(w, http.StatusInternalServerError, transferError("server_error", "temporary archive failure"))
		return
	}
	cmdtest.DefaultHandler(w, r)
}

func archivedMigrationFixtureHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	switch {
	case r.Method == http.MethodGet && p == "/projects/proj_cmdtest/apps":
		writeTransferJSON(w, http.StatusOK, transferList(transferApp("app_arch_source", "proj_cmdtest")))
	case r.Method == http.MethodGet && p == "/projects/proj_source/apps":
		writeTransferJSON(w, http.StatusOK, transferList(transferApp("app_source", "proj_source")))
	case r.Method == http.MethodGet && p == "/projects/proj_target/apps":
		writeTransferJSON(w, http.StatusOK, transferList(transferApp("app_target", "proj_target")))
	case r.Method == http.MethodGet && (p == "/projects/proj_cmdtest/products" || p == "/projects/proj_source/products"):
		writeTransferJSON(w, http.StatusOK, transferList(transferProduct("prod_source", "app_source", "archived")))
	case r.Method == http.MethodGet && p == "/projects/proj_target/products":
		writeTransferJSON(w, http.StatusOK, transferList(transferProduct("prod_target", "app_target", "active")))
	case r.Method == http.MethodGet && (p == "/projects/proj_cmdtest/entitlements" || p == "/projects/proj_source/entitlements"):
		writeTransferJSON(w, http.StatusOK, transferList(transferEntitlement("entl_source", "proj_source", "archived")))
	case r.Method == http.MethodGet && p == "/projects/proj_target/entitlements":
		writeTransferJSON(w, http.StatusOK, transferList(transferEntitlement("entl_target", "proj_target", "active")))
	case r.Method == http.MethodGet && strings.HasSuffix(p, "/entitlements/entl_source/products"):
		writeTransferJSON(w, http.StatusOK, transferList(transferProduct("prod_source", "app_source", "archived")))
	case r.Method == http.MethodGet && (p == "/projects/proj_cmdtest/offerings" || p == "/projects/proj_source/offerings"):
		writeTransferJSON(w, http.StatusOK, transferList(transferOffering("ofrnge_source", "proj_source", "archived")))
	case r.Method == http.MethodGet && p == "/projects/proj_target/offerings":
		writeTransferJSON(w, http.StatusOK, transferList(transferOffering("ofrnge_target", "proj_target", "active")))
	case r.Method == http.MethodGet && strings.HasSuffix(p, "/offerings/ofrnge_source/packages"):
		writeTransferJSON(w, http.StatusOK, transferList(transferPackage("pkge_source")))
	case r.Method == http.MethodGet && strings.HasSuffix(p, "/offerings/ofrnge_target/packages"):
		writeTransferJSON(w, http.StatusOK, transferList(transferPackage("pkge_target")))
	case r.Method == http.MethodGet && strings.HasSuffix(p, "/packages/pkge_source/products"):
		writeTransferJSON(w, http.StatusOK, transferList(map[string]any{
			"object":               "package_product",
			"product_id":           "prod_source",
			"eligibility_criteria": "all",
		}))
	default:
		writeTransferJSON(w, http.StatusNotFound, transferError("not_found", "no migration fixture for "+r.Method+" "+p))
	}
}

func newPartialImportFailureHandler() http.HandlerFunc {
	productCreateAttempts := 0
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		switch {
		case r.Method == http.MethodGet && p == "/projects/proj_cmdtest/apps":
			writeTransferJSON(w, http.StatusOK, transferList(transferApp("app_cmdtest", "proj_cmdtest")))
		case r.Method == http.MethodGet && (p == "/projects/proj_cmdtest/products" || p == "/projects/proj_cmdtest/entitlements" || p == "/projects/proj_cmdtest/offerings"):
			writeTransferJSON(w, http.StatusOK, transferList[map[string]any]())
		case r.Method == http.MethodPost && p == "/projects/proj_cmdtest/products":
			productCreateAttempts++
			if productCreateAttempts == 1 {
				writeTransferJSON(w, http.StatusInternalServerError, transferError("server_error", "temporary product create failure"))
				return
			}
			writeTransferJSON(w, http.StatusCreated, transferProduct("prod_good_target", "app_cmdtest", "active"))
		case r.Method == http.MethodPost && p == "/projects/proj_cmdtest/entitlements":
			writeTransferJSON(w, http.StatusCreated, transferEntitlement("entl_target", "proj_cmdtest", "active"))
		case r.Method == http.MethodPost && p == "/projects/proj_cmdtest/entitlements/entl_target/actions/attach_products":
			writeTransferJSON(w, http.StatusOK, map[string]any{"object": "entitlement", "id": "entl_target"})
		default:
			writeTransferJSON(w, http.StatusNotFound, transferError("not_found", "no partial-import fixture for "+r.Method+" "+p))
		}
	}
}

func writeTransferJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func transferList[T any](items ...T) map[string]any {
	return map[string]any{"object": "list", "items": items}
}

func transferError(typ, message string) map[string]any {
	return map[string]any{"object": "error", "type": typ, "message": message}
}

func transferApp(id, projectID string) map[string]any {
	return map[string]any{
		"object":     "app",
		"id":         id,
		"name":       "iOS App",
		"type":       "ios",
		"project_id": projectID,
		"created_at": 1713072000000,
	}
}

func transferProduct(id, appID, state string) map[string]any {
	return map[string]any{
		"object":           "product",
		"id":               id,
		"store_identifier": "com.example.premium.monthly",
		"type":             "subscription",
		"state":            state,
		"display_name":     "Premium Monthly",
		"app_id":           appID,
		"created_at":       1713072000000,
	}
}

func transferEntitlement(id, projectID, state string) map[string]any {
	return map[string]any{
		"object":       "entitlement",
		"id":           id,
		"project_id":   projectID,
		"lookup_key":   "premium",
		"display_name": "Premium",
		"state":        state,
		"created_at":   1713072000000,
	}
}

func transferOffering(id, projectID, state string) map[string]any {
	return map[string]any{
		"object":       "offering",
		"id":           id,
		"project_id":   projectID,
		"lookup_key":   "default",
		"display_name": "Default Offering",
		"is_current":   false,
		"state":        state,
		"created_at":   1713072000000,
		"metadata":     map[string]any{"tier": "gold"},
	}
}

func transferPackage(id string) map[string]any {
	position := 1
	return map[string]any{
		"object":       "package",
		"id":           id,
		"lookup_key":   "$rc_monthly",
		"display_name": "Monthly",
		"position":     position,
		"created_at":   1713072000000,
	}
}
