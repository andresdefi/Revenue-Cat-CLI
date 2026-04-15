package apps_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestAppsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "app_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps")
}

func TestAppsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps")
}

func TestAppsListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "apps", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "app_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps")
}

func TestAppsListProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "app_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/apps")
}

func TestAppsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestAppsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestAppsGetTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "get", "app_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "app_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps/app_cmdtest")
}

func TestAppsGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "get", "app_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"id\": \"app_cmdtest\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps/app_cmdtest")
}

func TestAppsGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestAppsCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "create", "--name", "iOS App", "--type", "app_store", "--bundle-id", "com.example.app", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "app_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/apps")
}

func TestAppsCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "create"})
	cmdtest.AssertErrorContains(t, result, "required flag")
}

func TestAppsDeleteSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "delete", "app_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "deleted")
	cmdtest.AssertRequested(t, result, "DELETE", "/projects/proj_cmdtest/apps/app_cmdtest")
}

func TestAppsDeleteMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "delete"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestAppsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestAppsHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestAppsUpdateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "update", "app_cmdtest", "--name", "Renamed App", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "app_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/apps/app_cmdtest")
}

func TestAppsUpdateMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "update"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestAppsUpdateAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "update", "app_cmdtest", "--name", "Renamed App", "--output", "json"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}
