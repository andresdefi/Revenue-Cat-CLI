package entitlements_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestEntitlementsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "entl_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/entitlements")
}

func TestEntitlementsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/entitlements")
}

func TestEntitlementsListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "entitlements", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "entl_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/entitlements")
}

func TestEntitlementsListProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "list", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "entl_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/entitlements")
}

func TestEntitlementsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestEntitlementsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestEntitlementsGetTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "get", "entl_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "entl_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/entitlements/entl_cmdtest")
}

func TestEntitlementsGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "get", "entl_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"id\": \"entl_cmdtest\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/entitlements/entl_cmdtest")
}

func TestEntitlementsGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestEntitlementsCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "create", "--lookup-key", "premium", "--display-name", "Premium", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "entl_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/entitlements")
}

func TestEntitlementsCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "create"})
	cmdtest.AssertErrorContains(t, result, "required")
}

func TestEntitlementsDeleteSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "delete", "entl_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "deleted")
	cmdtest.AssertRequested(t, result, "DELETE", "/projects/proj_cmdtest/entitlements/entl_cmdtest")
}

func TestEntitlementsDeleteMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "delete"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestEntitlementsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestEntitlementsHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestEntitlementsUpdateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "update", "entl_cmdtest", "--display-name", "Premium Plus", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "entl_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/entitlements/entl_cmdtest")
}

func TestEntitlementsUpdateMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "update"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestEntitlementsUpdateAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "update", "entl_cmdtest", "--display-name", "Premium Plus", "--output", "json"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestProductsTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "products", "entl_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "prod_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/entitlements/entl_cmdtest/products")
}

func TestAttachSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "attach", "--entitlement-id", "entl_cmdtest", "--product-id", "prod_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Attached")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/entitlements/entl_cmdtest/actions/attach_products")
}

func TestDetachSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"entitlements", "detach", "--entitlement-id", "entl_cmdtest", "--product-id", "prod_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Detached")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/entitlements/entl_cmdtest/actions/detach_products")
}
