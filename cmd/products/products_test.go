package products_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestProductsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "prod_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/products")
}

func TestProductsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"object": "list"`)
}

func TestProductsListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", cmdtest.TestProfile, "products", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "prod_cmdtest")
}

func TestProductsListProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/products")
}

func TestProductsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestProductsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list"}, cmdtest.WithAPIError(400, "parameter_error", "bad products request"))
	cmdtest.AssertErrorContains(t, result, "bad products request")
}

func TestProductsGetTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "get", "prod_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Premium Monthly")
}

func TestProductsGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "get", "prod_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"id": "prod_cmdtest"`)
}

func TestProductsGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestProductsCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "create", "--store-id", "com.example.premium.monthly", "--app-id", "app_cmdtest", "--type", "subscription", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Product created successfully")
}

func TestProductsCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "create", "--store-id", "com.example.premium.monthly"})
	cmdtest.AssertErrorContains(t, result, "missing required value")
}

func TestProductsUpdateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "update", "prod_cmdtest", "--display-name", "Premium Annual", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Product updated")
}

func TestProductsUpdateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "update", "prod_cmdtest"})
	cmdtest.AssertErrorContains(t, result, "required flag")
}

func TestProductsDeleteSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "delete", "prod_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "deleted")
}

func TestProductsDeleteMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "delete"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestProductsArchiveSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "archive", "prod_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "archived")
}

func TestProductsUnarchiveSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "unarchive", "prod_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "unarchived")
}

func TestProductsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestProductsHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}
