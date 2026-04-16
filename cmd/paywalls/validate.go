package paywalls

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

type validationReport struct {
	Object    string            `json:"object"`
	ProjectID string            `json:"project_id"`
	Status    string            `json:"status"`
	Counts    validationCounts  `json:"counts"`
	Checks    []validationCheck `json:"checks"`
}

type validationCounts struct {
	Paywalls            int `json:"paywalls"`
	Offerings           int `json:"offerings"`
	CurrentOfferings    int `json:"current_offerings"`
	Packages            int `json:"packages"`
	PackageProductLinks int `json:"package_product_links"`
}

type validationCheck struct {
	Status  string   `json:"status"`
	Area    string   `json:"area"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

func newValidateCmd(projectID, outputFormat *string) *cobra.Command {
	var strict bool
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate paywall readiness",
		Long: `Validate paywall readiness.

The validator is read-only. It checks that paywalls exist and that the current
offering has packages with product links, which are the RevenueCat paths a
paywall needs before launch.`,
		Example: `  # Validate paywall readiness
  rc paywalls validate

  # Emit JSON for automation
  rc paywalls validate --output json

  # Return non-zero when blocking checks fail
  rc paywalls validate --strict`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			report, err := validatePaywalls(client, pid)
			if err != nil {
				return err
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, report, renderValidationReport(report))
			if strict && report.Status == "fail" {
				return fmt.Errorf("paywall validation failed")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&strict, "strict", false, "return a non-zero exit code when failed checks are found")
	return cmd
}

func validatePaywalls(client *api.Client, projectID string) (*validationReport, error) {
	report := &validationReport{Object: "paywall_validation_report", ProjectID: projectID, Status: "pass"}

	paywalls, err := api.PaginateAll[api.Paywall](client, fmt.Sprintf("/projects/%s/paywalls", url.PathEscape(projectID)), nil)
	if err != nil {
		return nil, fmt.Errorf("list paywalls: %w", err)
	}
	report.Counts.Paywalls = len(paywalls)
	if len(paywalls) == 0 {
		report.add("fail", "paywalls", "No paywalls found in this project.")
	} else {
		report.add("pass", "paywalls", fmt.Sprintf("%d paywall(s) found.", len(paywalls)))
	}

	offerings, err := api.PaginateAll[api.Offering](client, fmt.Sprintf("/projects/%s/offerings", url.PathEscape(projectID)), nil)
	if err != nil {
		return nil, fmt.Errorf("list offerings: %w", err)
	}
	report.Counts.Offerings = len(offerings)
	current := currentActiveOfferings(offerings)
	report.Counts.CurrentOfferings = len(current)
	switch {
	case len(current) == 0:
		report.add("fail", "offering", "No current active offering is configured.")
		return report, nil
	case len(current) > 1:
		report.add("warn", "offering", fmt.Sprintf("%d current active offerings found; expected one.", len(current)))
	default:
		report.add("pass", "offering", fmt.Sprintf("Current offering is %s.", current[0].LookupKey))
	}

	packages, err := api.PaginateAll[api.Package](client, fmt.Sprintf("/projects/%s/offerings/%s/packages", url.PathEscape(projectID), url.PathEscape(current[0].ID)), nil)
	if err != nil {
		return nil, fmt.Errorf("list packages: %w", err)
	}
	report.Counts.Packages = len(packages)
	if len(packages) == 0 {
		report.add("fail", "packages", "Current offering has no packages.")
		return report, nil
	}

	details := make([]string, 0, len(packages))
	emptyPackages := make([]string, 0)
	for _, pkg := range packages {
		products, err := api.PaginateAll[api.PackageProduct](client, fmt.Sprintf("/projects/%s/packages/%s/products", url.PathEscape(projectID), url.PathEscape(pkg.ID)), nil)
		if err != nil {
			return nil, fmt.Errorf("list products for package %s: %w", pkg.ID, err)
		}
		report.Counts.PackageProductLinks += len(products)
		details = append(details, fmt.Sprintf("%s has %d product link(s)", pkg.LookupKey, len(products)))
		if len(products) == 0 {
			emptyPackages = append(emptyPackages, pkg.LookupKey)
		}
	}
	report.add("pass", "packages", fmt.Sprintf("%d package(s) found.", len(packages)), details...)
	if len(emptyPackages) > 0 {
		report.add("fail", "products", "Some packages have no product links.", emptyPackages...)
	} else {
		report.add("pass", "products", fmt.Sprintf("%d package-product link(s) found.", report.Counts.PackageProductLinks))
	}
	return report, nil
}

func currentActiveOfferings(offerings []api.Offering) []api.Offering {
	var current []api.Offering
	for _, offering := range offerings {
		if offering.IsCurrent && offering.State != "archived" {
			current = append(current, offering)
		}
	}
	return current
}

func (r *validationReport) add(status, area, message string, details ...string) {
	switch status {
	case "fail":
		r.Status = "fail"
	case "warn":
		if r.Status != "fail" {
			r.Status = "warn"
		}
	}
	r.Checks = append(r.Checks, validationCheck{Status: status, Area: area, Message: message, Details: details})
}

func renderValidationReport(report *validationReport) func(t table.Writer) {
	return func(t table.Writer) {
		t.AppendHeader(table.Row{"Status", "Area", "Message", "Details"})
		for _, check := range report.Checks {
			t.AppendRow(table.Row{strings.ToUpper(check.Status), check.Area, check.Message, strings.Join(check.Details, "\n")})
		}
		t.AppendFooter(table.Row{
			strings.ToUpper(report.Status),
			"paywalls",
			report.ProjectID,
			fmt.Sprintf("paywalls=%d current_offerings=%d packages=%d package_products=%d", report.Counts.Paywalls, report.Counts.CurrentOfferings, report.Counts.Packages, report.Counts.PackageProductLinks),
		})
	}
}
