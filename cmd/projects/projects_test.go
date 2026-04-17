package projects_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestProjectsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "proj_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestProjectsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestProjectsListAll(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "--all", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "proj_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestProjectsListLimit(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "--limit", "1", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "proj_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestProjectsListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "projects", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "proj_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestProjectsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestProjectsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestProjectsCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "create", "--name", "Command Test Project", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "proj_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects")
}

func TestProjectsCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "create"})
	cmdtest.AssertErrorContains(t, result, "missing required value: Project name")
}

func TestProjectsDoctorJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "doctor", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"object": "project_health_report"`)
	cmdtest.AssertOutputContains(t, result, `"status": "pass"`)
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/products")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/entitlements")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/packages/pkge_cmdtest/products")
}

func TestProjectDoctorAliasTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"project", "doctor", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "PASS")
	cmdtest.AssertOutputContains(t, result, "Current offering is default")
}

func TestProjectsDoctorReportsUnhealthyProject(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "doctor", "--output", "json"}, cmdtest.WithHandler(unhealthyProjectHandler))
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"status": "fail"`)
	cmdtest.AssertOutputContains(t, result, "No apps found")
	cmdtest.AssertOutputContains(t, result, "No current active offering")
}

func TestProjectsDoctorStrictFailsUnhealthyProject(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "doctor", "--strict"}, cmdtest.WithHandler(unhealthyProjectHandler))
	cmdtest.AssertErrorContains(t, result, "project doctor found errors")
}

func TestProjectsDoctorWatchRefreshesUntilContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := cmdtest.Run(t,
		[]string{"projects", "doctor", "--watch", "--interval", "1ns", "--output", "json"},
		cmdtest.WithContext(ctx),
		cmdtest.WithCancelOnRepeatedRequest(cancel),
	)

	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequestCountAtLeast(t, result, "GET", "/projects/proj_cmdtest/apps", 2)
}

func TestProjectsDoctorWatchAllowsNonPositiveInterval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := cmdtest.Run(t,
		[]string{"projects", "doctor", "--watch", "--interval", "0s", "--output", "json"},
		cmdtest.WithContext(ctx),
		cmdtest.WithHandler(func(w http.ResponseWriter, r *http.Request) {
			cancel()
			cmdtest.DefaultHandler(w, r)
		}),
	)

	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps")
}

func TestProjectsSetDefaultSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "set-default", "proj_new_default"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Default project set")
}

func TestProjectsSetDefaultMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "set-default"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestProjectsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestProjectsListHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestProjectsCreateHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "create", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestProjectsSetDefaultHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "set-default", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestProjectsRootHelp(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "projects")
}

func TestProjectsUnknownSubcommand(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "nope"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "projects")
}

func TestProjectsListShortOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "-o", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "proj_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func unhealthyProjectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeProjectTestJSON(w, http.StatusMethodNotAllowed, map[string]any{"object": "error", "type": "method_not_allowed", "message": "method not allowed"})
		return
	}

	switch p := r.URL.Path; {
	case strings.HasSuffix(p, "/apps"):
		writeProjectTestJSON(w, http.StatusOK, projectTestList())
	case strings.HasSuffix(p, "/products") && !strings.Contains(p, "/entitlements/") && !strings.Contains(p, "/packages/"):
		writeProjectTestJSON(w, http.StatusOK, projectTestList(projectTestProduct()))
	case strings.HasSuffix(p, "/entitlements"):
		writeProjectTestJSON(w, http.StatusOK, projectTestList(projectTestEntitlement()))
	case strings.Contains(p, "/entitlements/") && strings.HasSuffix(p, "/products"):
		writeProjectTestJSON(w, http.StatusOK, projectTestList())
	case strings.HasSuffix(p, "/offerings"):
		writeProjectTestJSON(w, http.StatusOK, projectTestList(projectTestOffering(false)))
	case strings.Contains(p, "/offerings/") && strings.HasSuffix(p, "/packages"):
		writeProjectTestJSON(w, http.StatusOK, projectTestList())
	default:
		writeProjectTestJSON(w, http.StatusNotFound, map[string]any{"object": "error", "type": "not_found", "message": "no fixture"})
	}
}

func projectTestList(items ...any) map[string]any {
	return map[string]any{"object": "list", "items": items, "next_page": nil}
}

func projectTestProduct() map[string]any {
	return map[string]any{"object": "product", "id": "prod_orphan", "store_identifier": "com.example.orphan", "type": "subscription", "state": "active", "app_id": "app_missing", "created_at": 1713072000000}
}

func projectTestEntitlement() map[string]any {
	return map[string]any{"object": "entitlement", "id": "entl_empty", "project_id": cmdtest.TestProjectID, "lookup_key": "premium", "display_name": "Premium", "state": "active", "created_at": 1713072000000}
}

func projectTestOffering(current bool) map[string]any {
	return map[string]any{"object": "offering", "id": "ofrnge_not_current", "project_id": cmdtest.TestProjectID, "lookup_key": "default", "display_name": "Default", "is_current": current, "state": "active", "created_at": 1713072000000}
}

func writeProjectTestJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
