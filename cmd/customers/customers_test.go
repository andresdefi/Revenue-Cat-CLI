package customers_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestCustomersListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "cust_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers")
}

func TestCustomersListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers")
}

func TestCustomersListSearch(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "list", "--search", "customer@example.com", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "cust_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers")
}

func TestCustomersLookupTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "lookup", "cust_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "cust_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest")
}

func TestCustomersLookupJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "lookup", "cust_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"id\": \"cust_cmdtest\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest")
}

func TestCustomersLookupMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "lookup"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestCustomersCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "create", "--id", "cust_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "cust_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/customers")
}

func TestCustomersCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "create"})
	cmdtest.AssertErrorContains(t, result, "missing required value: Customer ID")
}

func TestCustomersDeleteSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "delete", "cust_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "deleted")
	cmdtest.AssertRequested(t, result, "DELETE", "/projects/proj_cmdtest/customers/cust_cmdtest")
}

func TestCustomersEntitlementsJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "entitlements", "cust_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "entl_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest/active_entitlements")
}

func TestCustomersSubscriptionsJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "subscriptions", "cust_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "sub_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest/subscriptions")
}

func TestCustomersPurchasesJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "purchases", "cust_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "purch_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest/purchases")
}

func TestCustomersAliasesJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "aliases", "cust_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "alias_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest/aliases")
}

func TestCustomersAttributesJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "attributes", "cust_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "customer@example.com")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest/attributes")
}

func TestCustomersSetAttributesSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "set-attributes", "--customer-id", "cust_cmdtest", "--attr", "email=customer@example.com"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Attributes set")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/customers/cust_cmdtest/attributes")
}

func TestCustomersGrantSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "grant", "--customer-id", "cust_cmdtest", "--entitlement-id", "entl_cmdtest", "--expires-at", "1715750400000"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "granted")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/customers/cust_cmdtest/actions/grant_entitlement")
}

func TestCustomersRevokeSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "revoke", "--customer-id", "cust_cmdtest", "--entitlement-id", "entl_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "revoked")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/customers/cust_cmdtest/actions/revoke_granted_entitlement")
}

func TestCustomersInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}
