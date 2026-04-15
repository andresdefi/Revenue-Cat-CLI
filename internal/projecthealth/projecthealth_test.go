package projecthealth

import (
	"strings"
	"testing"

	"github.com/andresdefi/rc/internal/api"
)

func TestBuildReportPass(t *testing.T) {
	report := buildReport("proj_test", healthySnapshot())

	if report.Status != StatusPass {
		t.Fatalf("status = %q, want %q; checks = %#v", report.Status, StatusPass, report.Checks)
	}
	if report.Counts.Apps != 1 || report.Counts.ActiveProducts != 1 || report.Counts.CurrentOfferings != 1 {
		t.Fatalf("unexpected counts: %#v", report.Counts)
	}
}

func TestBuildReportWarnsForOrphanProduct(t *testing.T) {
	s := healthySnapshot()
	s.products = append(s.products, api.Product{
		ID:              "prod_orphan",
		StoreIdentifier: "com.example.orphan",
		State:           "active",
	})

	report := buildReport("proj_test", s)

	if report.Status != StatusWarn {
		t.Fatalf("status = %q, want %q; checks = %#v", report.Status, StatusWarn, report.Checks)
	}
	if !reportContains(report, "not attached to any entitlement") {
		t.Fatalf("report missing entitlement warning: %#v", report.Checks)
	}
	if !reportContains(report, "not attached to any package") {
		t.Fatalf("report missing package warning: %#v", report.Checks)
	}
}

func TestBuildReportFailsWithoutCurrentOffering(t *testing.T) {
	s := healthySnapshot()
	s.offerings[0].offering.IsCurrent = false

	report := buildReport("proj_test", s)

	if report.Status != StatusFail {
		t.Fatalf("status = %q, want %q; checks = %#v", report.Status, StatusFail, report.Checks)
	}
	if !reportContains(report, "No current active offering") {
		t.Fatalf("report missing current offering failure: %#v", report.Checks)
	}
}

func TestAssessLaunchPass(t *testing.T) {
	report := buildReport("proj_test", healthySnapshot())

	launch := AssessLaunch(report)

	if !launch.Ready {
		t.Fatalf("ready = false, want true; checks = %#v", launch.Checks)
	}
	if launch.Status != StatusPass {
		t.Fatalf("status = %q, want %q; checks = %#v", launch.Status, StatusPass, launch.Checks)
	}
}

func TestAssessLaunchWarnsButCanBeReady(t *testing.T) {
	report := buildReport("proj_test", healthySnapshot())
	report.add(Check{Status: StatusWarn, Area: "metadata", Message: "Optional warning"})

	launch := AssessLaunch(report)

	if !launch.Ready {
		t.Fatalf("ready = false, want true; checks = %#v", launch.Checks)
	}
	if launch.Status != StatusWarn {
		t.Fatalf("status = %q, want %q; checks = %#v", launch.Status, StatusWarn, launch.Checks)
	}
}

func TestAssessLaunchFailsWithoutRequiredPaths(t *testing.T) {
	report := buildReport("proj_test", &snapshot{})

	launch := AssessLaunch(report)

	if launch.Ready {
		t.Fatalf("ready = true, want false; checks = %#v", launch.Checks)
	}
	if launch.Status != StatusFail {
		t.Fatalf("status = %q, want %q; checks = %#v", launch.Status, StatusFail, launch.Checks)
	}
	if !launchContains(launch, "At least one app is required") {
		t.Fatalf("launch report missing app failure: %#v", launch.Checks)
	}
}

func healthySnapshot() *snapshot {
	product := api.Product{
		ID:              "prod_monthly",
		StoreIdentifier: "com.example.monthly",
		State:           "active",
	}
	return &snapshot{
		apps: []api.App{{ID: "app_ios", Name: "iOS", Type: "app_store"}},
		products: []api.Product{
			product,
		},
		entitlements: []entitlementProducts{
			{
				entitlement: api.Entitlement{
					ID:        "entl_premium",
					LookupKey: "premium",
					State:     "active",
				},
				products: []api.Product{product},
			},
		},
		offerings: []offeringPackages{
			{
				offering: api.Offering{
					ID:        "ofrnge_default",
					LookupKey: "default",
					IsCurrent: true,
					State:     "active",
				},
				packages: []packageProducts{
					{
						pkg: api.Package{
							ID:        "pkge_monthly",
							LookupKey: "$rc_monthly",
						},
						products: []api.PackageProduct{{ProductID: product.ID}},
					},
				},
			},
		},
	}
}

func reportContains(report *Report, want string) bool {
	for _, check := range report.Checks {
		if strings.Contains(check.Message, want) {
			return true
		}
		for _, detail := range check.Details {
			if strings.Contains(detail, want) {
				return true
			}
		}
	}
	return false
}

func launchContains(report *LaunchReport, want string) bool {
	for _, check := range report.Checks {
		if strings.Contains(check.Message, want) {
			return true
		}
		for _, detail := range check.Details {
			if strings.Contains(detail, want) {
				return true
			}
		}
	}
	return false
}
