package paywalls_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestPaywallsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "paywall_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/paywalls")
}

func TestPaywallsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/paywalls")
}

func TestPaywallsListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "paywalls", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "paywall_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/paywalls")
}

func TestPaywallsListProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "list", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "paywall_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/paywalls")
}

func TestPaywallsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestPaywallsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestPaywallsGetTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "get", "paywall_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "paywall_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/paywalls/paywall_cmdtest")
}

func TestPaywallsGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "get", "paywall_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"id\": \"paywall_cmdtest\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/paywalls/paywall_cmdtest")
}

func TestPaywallsGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestPaywallsCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "create", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "paywall_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/paywalls")
}

func TestPaywallsCreateTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "create"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "paywall_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/paywalls")
}

func TestPaywallsDeleteSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "delete", "paywall_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "deleted")
	cmdtest.AssertRequested(t, result, "DELETE", "/projects/proj_cmdtest/paywalls/paywall_cmdtest")
}

func TestPaywallsDeleteMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "delete"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestPaywallsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestPaywallsHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestPaywallsCreateAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "create", "--output", "json"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestPaywallsListAll(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "list", "--all", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "paywall_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/paywalls")
}

func TestPaywallsListLimit(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "list", "--limit", "1", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "paywall_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/paywalls")
}
