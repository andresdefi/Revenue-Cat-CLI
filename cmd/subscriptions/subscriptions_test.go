package subscriptions_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestSubscriptionsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "sub_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/subscriptions")
}

func TestSubscriptionsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/subscriptions")
}

func TestSubscriptionsListAll(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "list", "--all", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "sub_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/subscriptions")
}

func TestSubscriptionsGetTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "get", "sub_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "sub_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/subscriptions/sub_cmdtest")
}

func TestSubscriptionsGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "get", "sub_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"id\": \"sub_cmdtest\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/subscriptions/sub_cmdtest")
}

func TestSubscriptionsGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestSubscriptionsTransactionsJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "transactions", "sub_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "txn_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/subscriptions/sub_cmdtest/transactions")
}

func TestSubscriptionsEntitlementsJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "entitlements", "sub_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "entl_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/subscriptions/sub_cmdtest/entitlements")
}

func TestSubscriptionsCancelSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "cancel", "sub_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "canceled")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/subscriptions/sub_cmdtest/actions/cancel")
}

func TestSubscriptionsRefundSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "refund", "sub_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "refunded")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/subscriptions/sub_cmdtest/actions/refund")
}

func TestSubscriptionsRefundTransactionSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "refund-transaction", "sub_cmdtest", "--transaction-id", "txn_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "refunded")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/subscriptions/sub_cmdtest/transactions/txn_cmdtest/actions/refund")
}

func TestSubscriptionsRefundTransactionMissingFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "refund-transaction", "sub_cmdtest"})
	cmdtest.AssertErrorContains(t, result, "required flag")
}

func TestSubscriptionsManagementURLJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "management-url", "sub_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "https://pay.rev.cat/manage")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/subscriptions/sub_cmdtest/authenticated_management_url")
}

func TestSubscriptionsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestSubscriptionsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestSubscriptionsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestSubscriptionsHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestSubscriptionsRootHelp(t *testing.T) {
	result := cmdtest.Run(t, []string{"subscriptions", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "subscriptions")
}
