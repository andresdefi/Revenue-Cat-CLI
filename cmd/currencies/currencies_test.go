package currencies_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestCurrenciesListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "COIN")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/virtual_currencies")
}

func TestCurrenciesListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/virtual_currencies")
}

func TestCurrenciesGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "get", "COIN", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"code\": \"COIN\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/virtual_currencies/COIN")
}

func TestCurrenciesGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestCurrenciesCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "create", "--code", "COIN", "--name", "Coins", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "COIN")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/virtual_currencies")
}

func TestCurrenciesCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "create"})
	cmdtest.AssertErrorContains(t, result, "missing required value: Currency code")
}

func TestCurrenciesCreateMissingNameFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "create", "--code", "COIN"})
	cmdtest.AssertErrorContains(t, result, "missing required value: Display name")
}

func TestCurrenciesUpdateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "update", "COIN", "--name", "Coins Plus", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "COIN")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/virtual_currencies/COIN")
}

func TestCurrenciesUpdateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "update", "COIN"})
	cmdtest.AssertErrorContains(t, result, "required flag")
}

func TestCurrenciesDeleteSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "delete", "COIN"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "deleted")
	cmdtest.AssertRequested(t, result, "DELETE", "/projects/proj_cmdtest/virtual_currencies/COIN")
}

func TestCurrenciesArchiveSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "archive", "COIN"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "archived")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/virtual_currencies/COIN/actions/archive")
}

func TestCurrenciesUnarchiveSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "unarchive", "COIN"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "unarchived")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/virtual_currencies/COIN/actions/unarchive")
}

func TestCurrenciesBalanceJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "balance", "--customer-id", "cust_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "COIN")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest/virtual_currencies")
}

func TestCurrenciesCreditSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "credit", "--customer-id", "cust_cmdtest", "--code", "COIN", "--amount", "100"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Transaction created")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/customers/cust_cmdtest/virtual_currencies/transactions")
	cmdtest.AssertRequestJSON(t, result, "POST", "/projects/proj_cmdtest/customers/cust_cmdtest/virtual_currencies/transactions", map[string]any{
		"adjustments": map[string]any{"COIN": float64(100)},
	})
}

func TestCurrenciesSetBalanceSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "set-balance", "--customer-id", "cust_cmdtest", "--code", "COIN", "--balance", "500"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Balance set")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/customers/cust_cmdtest/virtual_currencies/update_balance")
	cmdtest.AssertRequestJSON(t, result, "POST", "/projects/proj_cmdtest/customers/cust_cmdtest/virtual_currencies/update_balance", map[string]any{
		"adjustments": map[string]any{"COIN": float64(500)},
	})
}

func TestCurrenciesListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestCurrenciesListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestCurrenciesInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestCurrenciesHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"currencies", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}
