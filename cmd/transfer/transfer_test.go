package transfer_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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
