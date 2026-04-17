package paywalls_test

import (
	"encoding/json"
	"net/http"
	"strings"
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
	result := cmdtest.Run(t, []string{"paywalls", "create", "--offering-id", "ofrnge_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "paywall_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/paywalls")
}

func TestPaywallsCreateTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "create", "--offering-id", "ofrnge_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "paywall_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/paywalls")
}

func TestPaywallsCreateMissingOfferingID(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "create"})
	cmdtest.AssertErrorContains(t, result, "required flag")
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
	result := cmdtest.Run(t, []string{"paywalls", "create", "--offering-id", "ofrnge_cmdtest", "--output", "json"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
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

func TestPaywallsValidatePasses(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "validate", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"object": "paywall_validation_report"`)
	cmdtest.AssertOutputContains(t, result, `"status": "pass"`)
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_cmdtest/paywalls")
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_cmdtest/offerings")
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/packages")
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_cmdtest/packages/pkge_cmdtest/products")
}

func TestPaywallsValidateStrictFails(t *testing.T) {
	result := cmdtest.Run(t, []string{"paywalls", "validate", "--strict"}, cmdtest.WithHandler(emptyPaywallValidationHandler))
	cmdtest.AssertErrorContains(t, result, "paywall validation failed")
	cmdtest.AssertOutputContains(t, result, "No paywalls found")
}

func emptyPaywallValidationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet && (strings.HasSuffix(r.URL.Path, "/paywalls") || strings.HasSuffix(r.URL.Path, "/offerings")) {
		writePaywallJSON(w, paywallList())
		return
	}
	writePaywallJSON(w, map[string]any{"object": "error", "type": "not_found", "message": "not found"})
}

func paywallList(items ...any) map[string]any {
	return map[string]any{"object": "list", "items": items, "next_page": nil, "url": "/fixture"}
}

func writePaywallJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
