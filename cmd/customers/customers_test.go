package customers_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

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

func TestCustomerAccessViewsWatchRefreshUntilContextCanceled(t *testing.T) {
	tests := []struct {
		name string
		args []string
		path string
	}{
		{
			name: "lookup",
			args: []string{"customers", "lookup", "cust_cmdtest", "--watch", "--interval", "1ns", "--output", "json"},
			path: "/projects/proj_cmdtest/customers/cust_cmdtest",
		},
		{
			name: "entitlements",
			args: []string{"customers", "entitlements", "cust_cmdtest", "--watch", "--interval", "1ns", "--output", "json"},
			path: "/projects/proj_cmdtest/customers/cust_cmdtest/active_entitlements",
		},
		{
			name: "subscriptions",
			args: []string{"customers", "subscriptions", "cust_cmdtest", "--watch", "--interval", "1ns", "--output", "json"},
			path: "/projects/proj_cmdtest/customers/cust_cmdtest/subscriptions",
		},
		{
			name: "purchases",
			args: []string{"customers", "purchases", "cust_cmdtest", "--watch", "--interval", "1ns", "--output", "json"},
			path: "/projects/proj_cmdtest/customers/cust_cmdtest/purchases",
		},
		{
			name: "aliases",
			args: []string{"customers", "aliases", "cust_cmdtest", "--watch", "--interval", "1ns", "--output", "json"},
			path: "/projects/proj_cmdtest/customers/cust_cmdtest/aliases",
		},
		{
			name: "attributes",
			args: []string{"customers", "attributes", "cust_cmdtest", "--watch", "--interval", "1ns", "--output", "json"},
			path: "/projects/proj_cmdtest/customers/cust_cmdtest/attributes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			result := cmdtest.Run(t,
				tt.args,
				cmdtest.WithContext(ctx),
				cmdtest.WithCancelOnRepeatedRequest(cancel),
			)

			cmdtest.AssertSuccess(t, result)
			cmdtest.AssertRequestCountAtLeast(t, result, "GET", tt.path, 2)
		})
	}
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

func TestCustomersGrantDurationComputesExpiration(t *testing.T) {
	before := time.Now().Add(29 * 24 * time.Hour).UnixMilli()
	result := cmdtest.Run(t, []string{"customers", "grant", "--customer-id", "cust_cmdtest", "--entitlement-id", "entl_cmdtest", "--duration", "30d"})
	after := time.Now().Add(31 * 24 * time.Hour).UnixMilli()

	cmdtest.AssertSuccess(t, result)
	body := customerGrantBody(t, result)
	if body["entitlement_id"] != "entl_cmdtest" {
		t.Fatalf("entitlement_id = %#v, want entl_cmdtest", body["entitlement_id"])
	}
	expiresAt, ok := body["expires_at"].(float64)
	if !ok {
		t.Fatalf("expires_at = %#v, want number", body["expires_at"])
	}
	if got := int64(expiresAt); got < before || got > after {
		t.Fatalf("expires_at = %d, want between %d and %d", got, before, after)
	}
}

func TestCustomersGrantLifetimeUsesFarFutureExpiration(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "grant", "--customer-id", "cust_cmdtest", "--entitlement-id", "entl_cmdtest", "--lifetime"})

	cmdtest.AssertSuccess(t, result)
	body := customerGrantBody(t, result)
	if got := int64(body["expires_at"].(float64)); got != 253402300799999 {
		t.Fatalf("expires_at = %d, want far-future lifetime timestamp", got)
	}
}

func TestCustomersGrantRequiresExpirationChoice(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "grant", "--customer-id", "cust_cmdtest", "--entitlement-id", "entl_cmdtest"})
	cmdtest.AssertErrorContains(t, result, "one of --expires-at, --duration, or --lifetime is required")
}

func TestCustomersGrantRejectsMultipleExpirationChoices(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "grant", "--customer-id", "cust_cmdtest", "--entitlement-id", "entl_cmdtest", "--duration", "30d", "--lifetime"})
	cmdtest.AssertErrorContains(t, result, "use only one of --expires-at, --duration, or --lifetime")
}

func TestCustomersGrantRejectsInvalidDuration(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "grant", "--customer-id", "cust_cmdtest", "--entitlement-id", "entl_cmdtest", "--duration", "soon"})
	cmdtest.AssertErrorContains(t, result, "invalid --duration")
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

func customerGrantBody(t *testing.T, result cmdtest.Result) map[string]any {
	t.Helper()
	for _, req := range result.Requests {
		if req.Method == "POST" && req.Path == "/projects/proj_cmdtest/customers/cust_cmdtest/actions/grant_entitlement" {
			var body map[string]any
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				t.Fatalf("grant body is not JSON: %v", err)
			}
			return body
		}
	}
	t.Fatalf("missing grant request; got %#v", result.Requests)
	return nil
}
