package customers_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestCustomersDiagnoseActiveEntitlementPasses(t *testing.T) {
	result := cmdtest.Run(t, []string{"customers", "diagnose", "cust_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest/active_entitlements")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest/subscriptions")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest/purchases")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest/aliases")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/customers/cust_cmdtest/attributes")

	var report diagnosisReport
	mustDecodeDiagnosis(t, result.Stdout, &report)
	if report.Object != "customer_diagnosis" {
		t.Fatalf("object = %q, want customer_diagnosis", report.Object)
	}
	if report.AccessSummary != "has_access" {
		t.Fatalf("access_summary = %q, want has_access", report.AccessSummary)
	}
	if report.Status != "pass" {
		t.Fatalf("status = %q, want pass", report.Status)
	}
	if len(report.ActiveEntitlements) != 1 || report.ActiveEntitlements[0].EntitlementID != "entl_cmdtest" {
		t.Fatalf("active_entitlements = %#v, want entl_cmdtest", report.ActiveEntitlements)
	}
}

func TestCustomersDiagnoseSubscriptionWithoutEntitlementFails(t *testing.T) {
	result := cmdtest.Run(t,
		[]string{"customers", "diagnose", "cust_cmdtest", "--output", "json"},
		cmdtest.WithHandler(diagnoseFixture(diagnoseScenario{
			entitlements: []any{},
			subscriptions: []any{
				subscriptionFixture("sub_no_access", "prod_unlinked", "active", false),
			},
			purchases:  []any{},
			aliases:    []any{},
			attributes: []any{},
		})),
	)
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Active subscription(s) are present but do not line up with active entitlement access.")

	var report diagnosisReport
	mustDecodeDiagnosis(t, result.Stdout, &report)
	if report.AccessSummary != "no_access" {
		t.Fatalf("access_summary = %q, want no_access", report.AccessSummary)
	}
	if report.Status != "fail" {
		t.Fatalf("status = %q, want fail", report.Status)
	}
}

func TestCustomersDiagnoseNoHistoryReportsNoAccess(t *testing.T) {
	result := cmdtest.Run(t,
		[]string{"customers", "diagnose", "cust_empty", "--output", "table"},
		cmdtest.WithHandler(diagnoseFixture(diagnoseScenario{
			customerID:    "cust_empty",
			entitlements:  []any{},
			subscriptions: []any{},
			purchases:     []any{},
			aliases:       []any{},
			attributes:    []any{},
		})),
	)
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "NO_ACCESS")
	cmdtest.AssertOutputContains(t, result, "No purchases or subscriptions were found for this customer.")
}

func TestCustomersDiagnoseSurfacesAliases(t *testing.T) {
	result := cmdtest.Run(t,
		[]string{"customers", "diagnose", "cust_cmdtest", "--output", "json"},
		cmdtest.WithHandler(diagnoseFixture(diagnoseScenario{
			entitlements: []any{activeEntitlementFixture("entl_cmdtest")},
			subscriptions: []any{
				subscriptionFixture("sub_cmdtest", "prod_cmdtest", "active", true),
			},
			purchases: []any{},
			aliases: []any{
				map[string]any{"object": "customer_alias", "id": "alias_primary"},
				map[string]any{"object": "customer_alias", "id": "alias_secondary"},
			},
			attributes: []any{},
		})),
	)
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "alias_secondary")

	var report diagnosisReport
	mustDecodeDiagnosis(t, result.Stdout, &report)
	if len(report.Aliases) != 2 {
		t.Fatalf("aliases = %#v, want two aliases", report.Aliases)
	}
}

func TestCustomersDiagnoseStrictReturnsNonZeroOnFailedChecks(t *testing.T) {
	result := cmdtest.Run(t,
		[]string{"customers", "diagnose", "cust_cmdtest", "--strict"},
		cmdtest.WithHandler(diagnoseFixture(diagnoseScenario{
			entitlements: []any{},
			subscriptions: []any{
				subscriptionFixture("sub_no_access", "prod_unlinked", "active", false),
			},
			purchases:  []any{},
			aliases:    []any{},
			attributes: []any{},
		})),
	)
	cmdtest.AssertErrorContains(t, result, "customer diagnosis found failed checks")
}

func TestCustomersDiagnoseJSONShape(t *testing.T) {
	result := cmdtest.Run(t, []string{"customer", "diagnose", "cust_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)

	var report diagnosisReport
	mustDecodeDiagnosis(t, result.Stdout, &report)
	if report.CustomerID != "cust_cmdtest" {
		t.Fatalf("customer_id = %q, want cust_cmdtest", report.CustomerID)
	}
	if report.Counts.ActiveEntitlements != 1 {
		t.Fatalf("counts.active_entitlements = %d, want 1", report.Counts.ActiveEntitlements)
	}
	if len(report.Subscriptions) != 1 || report.Subscriptions[0].ProductID != "prod_cmdtest" {
		t.Fatalf("subscriptions = %#v, want prod_cmdtest summary", report.Subscriptions)
	}
	if len(report.NextCommands) == 0 || !strings.Contains(report.NextCommands[0], "rc customers entitlements cust_cmdtest") {
		t.Fatalf("next_commands = %#v, want customer entitlement command first", report.NextCommands)
	}
}

type diagnosisReport struct {
	Object        string `json:"object"`
	ProjectID     string `json:"project_id"`
	CustomerID    string `json:"customer_id"`
	AccessSummary string `json:"access_summary"`
	Status        string `json:"status"`
	Counts        struct {
		ActiveEntitlements int `json:"active_entitlements"`
		Subscriptions      int `json:"subscriptions"`
		Purchases          int `json:"purchases"`
		Aliases            int `json:"aliases"`
	} `json:"counts"`
	ActiveEntitlements []struct {
		EntitlementID string `json:"entitlement_id"`
	} `json:"active_entitlements"`
	Subscriptions []struct {
		ID          string `json:"id"`
		ProductID   string `json:"product_id"`
		Status      string `json:"status"`
		GivesAccess bool   `json:"gives_access"`
	} `json:"subscriptions"`
	Aliases      []string `json:"aliases"`
	NextCommands []string `json:"next_commands"`
}

func mustDecodeDiagnosis(t *testing.T, stdout string, report *diagnosisReport) {
	t.Helper()
	if err := json.Unmarshal([]byte(stdout), report); err != nil {
		t.Fatalf("failed to decode diagnosis JSON: %v\nstdout:\n%s", err, stdout)
	}
}

type diagnoseScenario struct {
	customerID    string
	entitlements  []any
	subscriptions []any
	purchases     []any
	aliases       []any
	attributes    []any
}

func diagnoseFixture(s diagnoseScenario) http.HandlerFunc {
	if s.customerID == "" {
		s.customerID = "cust_cmdtest"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		switch {
		case strings.HasSuffix(p, "/active_entitlements"):
			writeDiagnosisJSON(w, listFixture(s.entitlements...))
		case strings.HasSuffix(p, "/subscriptions"):
			writeDiagnosisJSON(w, listFixture(s.subscriptions...))
		case strings.HasSuffix(p, "/purchases"):
			writeDiagnosisJSON(w, listFixture(s.purchases...))
		case strings.HasSuffix(p, "/aliases"):
			writeDiagnosisJSON(w, listFixture(s.aliases...))
		case strings.HasSuffix(p, "/attributes"):
			writeDiagnosisJSON(w, listFixture(s.attributes...))
		case strings.Contains(p, "/customers/"):
			writeDiagnosisJSON(w, map[string]any{
				"object":              "customer",
				"id":                  s.customerID,
				"project_id":          cmdtest.TestProjectID,
				"first_seen_at":       1713072000000,
				"active_entitlements": listFixture(s.entitlements...),
			})
		default:
			writeDiagnosisJSON(w, map[string]any{"object": "error", "type": "not_found", "message": "not found"})
		}
	}
}

func listFixture(items ...any) map[string]any {
	return map[string]any{
		"object":    "list",
		"items":     items,
		"next_page": nil,
		"url":       "/fixture",
	}
}

func activeEntitlementFixture(entitlementID string) map[string]any {
	return map[string]any{"object": "active_entitlement", "entitlement_id": entitlementID, "expires_at": nil}
}

func subscriptionFixture(id, productID, status string, givesAccess bool) map[string]any {
	return map[string]any{
		"object":                        "subscription",
		"id":                            id,
		"customer_id":                   "cust_cmdtest",
		"original_customer_id":          "cust_cmdtest",
		"product_id":                    productID,
		"starts_at":                     1713072000000,
		"current_period_starts_at":      1713072000000,
		"current_period_ends_at":        1715750400000,
		"ends_at":                       nil,
		"gives_access":                  givesAccess,
		"pending_payment":               false,
		"auto_renewal_status":           "will_renew",
		"status":                        status,
		"presented_offering_id":         "ofrnge_cmdtest",
		"environment":                   "production",
		"store":                         "app_store",
		"store_subscription_identifier": "store_sub_cmdtest",
		"ownership":                     "purchased",
		"country":                       "US",
	}
}

func writeDiagnosisJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
