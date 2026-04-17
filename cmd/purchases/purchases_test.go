package purchases_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestPurchasesListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "list", "--store-purchase-id", "store_purchase_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "purch_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/purchases")
}

func TestPurchasesListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "list", "--store-purchase-id", "store_purchase_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/purchases")
	cmdtest.AssertRequestedWithQuery(t, result, "GET", "/projects/proj_cmdtest/purchases", "store_purchase_identifier", "store_purchase_cmdtest")
}

func TestPurchasesListAll(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "list", "--store-purchase-id", "store_purchase_cmdtest", "--all", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "purch_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/purchases")
}

func TestPurchasesGetTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "get", "purch_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "purch_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/purchases/purch_cmdtest")
}

func TestPurchasesGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "get", "purch_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"id\": \"purch_cmdtest\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/purchases/purch_cmdtest")
}

func TestPurchasesGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestPurchasesEntitlementsJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "entitlements", "purch_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "entl_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/purchases/purch_cmdtest/entitlements")
}

func TestPurchasesRefundSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "refund", "purch_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "refunded")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/purchases/purch_cmdtest/actions/refund")
}

func TestPurchasesRefundMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "refund"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestPurchasesListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "purchases", "list", "--store-purchase-id", "store_purchase_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "purch_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/purchases")
}

func TestPurchasesProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "list", "--store-purchase-id", "store_purchase_cmdtest", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "purch_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/purchases")
}

func TestPurchasesListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "list", "--store-purchase-id", "store_purchase_cmdtest"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestPurchasesListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "list", "--store-purchase-id", "store_purchase_cmdtest"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestPurchasesInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "list", "--store-purchase-id", "store_purchase_cmdtest", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestPurchasesHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "list", "--store-purchase-id", "store_purchase_cmdtest", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestPurchasesRootHelp(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "purchases")
}

func TestPurchasesListLimit(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "list", "--store-purchase-id", "store_purchase_cmdtest", "--limit", "1", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "purch_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/purchases")
}

func TestPurchasesEntitlementsMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"purchases", "entitlements"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}
