package packages_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestPackagesListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "list", "--offering-id", "ofrnge_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "pkge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/packages")
}

func TestPackagesListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "list", "--offering-id", "ofrnge_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/packages")
}

func TestPackagesListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "packages", "list", "--offering-id", "ofrnge_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "pkge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/packages")
}

func TestPackagesListProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "list", "--offering-id", "ofrnge_cmdtest", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "pkge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/offerings/ofrnge_cmdtest/packages")
}

func TestPackagesListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "list", "--offering-id", "ofrnge_cmdtest"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestPackagesListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "list", "--offering-id", "ofrnge_cmdtest"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestPackagesGetTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "get", "pkge_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "pkge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/packages/pkge_cmdtest")
}

func TestPackagesGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "get", "pkge_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"id\": \"pkge_cmdtest\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/packages/pkge_cmdtest")
}

func TestPackagesGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestPackagesCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "create", "--offering-id", "ofrnge_cmdtest", "--lookup-key", "$rc_monthly", "--display-name", "Monthly", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "pkge_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/packages")
}

func TestPackagesCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "create"})
	cmdtest.AssertErrorContains(t, result, "missing required value: Offering ID")
}

func TestPackagesCreateMissingLookupKeyFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "create", "--offering-id", "ofrnge_cmdtest"})
	cmdtest.AssertErrorContains(t, result, "missing required value: Lookup key")
}

func TestPackagesCreateMissingDisplayNameFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "create", "--offering-id", "ofrnge_cmdtest", "--lookup-key", "$rc_monthly"})
	cmdtest.AssertErrorContains(t, result, "missing required value: Display name")
}

func TestPackagesDeleteSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "delete", "pkge_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "deleted")
	cmdtest.AssertRequested(t, result, "DELETE", "/projects/proj_cmdtest/packages/pkge_cmdtest")
}

func TestPackagesDeleteMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "delete"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestPackagesInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "list", "--offering-id", "ofrnge_cmdtest", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestPackagesHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestPackagesUpdateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "update", "pkge_cmdtest", "--display-name", "Monthly Plus", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "pkge_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/packages/pkge_cmdtest")
}

func TestPackagesUpdateMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "update"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestPackagesUpdateAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "update", "pkge_cmdtest", "--display-name", "Monthly Plus", "--output", "json"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestListWithOffering(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "list", "--offering-id", "ofrnge_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "pkge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest/packages")
}

func TestProductsTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "products", "pkge_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "prod_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/packages/pkge_cmdtest/products")
}

func TestAttachSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "attach", "--package-id", "pkge_cmdtest", "--product-id", "prod_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "attached")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/packages/pkge_cmdtest/actions/attach_products")
}

func TestDetachSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"packages", "detach", "--package-id", "pkge_cmdtest", "--product-id", "prod_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Detached")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/packages/pkge_cmdtest/actions/detach_products")
}
