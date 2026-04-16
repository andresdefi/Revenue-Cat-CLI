package offerings

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

type publishReport struct {
	Object     string         `json:"object"`
	ProjectID  string         `json:"project_id"`
	OfferingID string         `json:"offering_id"`
	DryRun     bool           `json:"dry_run"`
	Published  bool           `json:"published"`
	Status     string         `json:"status"`
	Checks     []publishCheck `json:"checks"`
}

type publishCheck struct {
	Status  string   `json:"status"`
	Area    string   `json:"area"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

type packageProductSet struct {
	pkg      api.Package
	products []api.PackageProduct
}

func newPublishCmd(projectID, outputFormat *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish <offering-id>",
		Short: "Validate and make an offering current",
		Long: `Validate and make an offering current.

Publish checks that the offering is active, has packages, and that each package
has product links before setting it as the current offering.`,
		Example: `  # Publish an offering after validation
  rc offerings publish ofrnge1a2b3c4d5

  # Preview the checks and publish result as JSON
  rc offerings publish ofrnge1a2b3c4d5 --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			report, err := publishOffering(client, pid, args[0])
			if err != nil {
				return err
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, report, renderPublishReport(report))
			if report.Status == "fail" {
				return fmt.Errorf("offering publish validation failed")
			}
			if report.Published {
				output.Success("Offering %s published", args[0])
			}
			return nil
		},
	}
	return cmd
}

func publishOffering(client *api.Client, projectID, offeringID string) (*publishReport, error) {
	report, offering, packages, err := buildPublishReport(client, projectID, offeringID)
	if err != nil {
		return nil, err
	}
	if report.Status == "fail" {
		return report, nil
	}
	if offering.IsCurrent {
		report.Published = true
		report.add("pass", "publish", "Offering is already current.")
		return report, nil
	}
	if api.DryRun {
		report.add("info", "publish", fmt.Sprintf("Would make offering %s current.", offeringID))
		return report, nil
	}
	if len(packages) == 0 {
		return report, nil
	}
	_, err = client.Post(fmt.Sprintf("/projects/%s/offerings/%s", url.PathEscape(projectID), url.PathEscape(offeringID)), map[string]any{"is_current": true})
	if err != nil {
		return nil, fmt.Errorf("publish offering: %w", err)
	}
	report.Published = true
	report.add("pass", "publish", fmt.Sprintf("Offering %s is now current.", offeringID))
	return report, nil
}

func buildPublishReport(client *api.Client, projectID, offeringID string) (*publishReport, api.Offering, []packageProductSet, error) {
	report := &publishReport{
		Object:     "offering_publish_report",
		ProjectID:  projectID,
		OfferingID: offeringID,
		DryRun:     api.DryRun,
		Status:     "pass",
	}

	data, err := client.Get(fmt.Sprintf("/projects/%s/offerings/%s", url.PathEscape(projectID), url.PathEscape(offeringID)), nil)
	if err != nil {
		return nil, api.Offering{}, nil, fmt.Errorf("get offering: %w", err)
	}
	var offering api.Offering
	if err := json.Unmarshal(data, &offering); err != nil {
		return nil, api.Offering{}, nil, fmt.Errorf("parse offering: %w", err)
	}

	if offering.State == "archived" {
		report.add("fail", "offering", "Offering is archived and cannot be published.")
	} else {
		report.add("pass", "offering", fmt.Sprintf("Offering %s is %s.", offering.LookupKey, valueOr(offering.State, "available")))
	}

	packages, err := fetchOfferingPackages(client, projectID, offeringID)
	if err != nil {
		return nil, api.Offering{}, nil, err
	}
	if len(packages) == 0 {
		report.add("fail", "packages", "Offering has no packages.")
		return report, offering, packages, nil
	}

	packageDetails := make([]string, 0, len(packages))
	emptyPackages := make([]string, 0)
	for _, pkg := range packages {
		packageDetails = append(packageDetails, fmt.Sprintf("%s has %d product link(s)", pkg.pkg.LookupKey, len(pkg.products)))
		if len(pkg.products) == 0 {
			emptyPackages = append(emptyPackages, pkg.pkg.LookupKey)
		}
	}
	report.add("pass", "packages", fmt.Sprintf("%d package(s) found.", len(packages)), packageDetails...)
	if len(emptyPackages) > 0 {
		report.add("fail", "products", "Some packages have no product links.", emptyPackages...)
	} else {
		report.add("pass", "products", "Every package has at least one product link.")
	}
	return report, offering, packages, nil
}

func fetchOfferingPackages(client *api.Client, projectID, offeringID string) ([]packageProductSet, error) {
	packages, err := api.PaginateAll[api.Package](client, fmt.Sprintf("/projects/%s/offerings/%s/packages", url.PathEscape(projectID), url.PathEscape(offeringID)), nil)
	if err != nil {
		return nil, fmt.Errorf("list offering packages: %w", err)
	}
	result := make([]packageProductSet, 0, len(packages))
	for _, pkg := range packages {
		products, err := api.PaginateAll[api.PackageProduct](client, fmt.Sprintf("/projects/%s/packages/%s/products", url.PathEscape(projectID), url.PathEscape(pkg.ID)), nil)
		if err != nil {
			return nil, fmt.Errorf("list products for package %s: %w", pkg.ID, err)
		}
		result = append(result, packageProductSet{pkg: pkg, products: products})
	}
	return result, nil
}

func (r *publishReport) add(status, area, message string, details ...string) {
	if status == "fail" {
		r.Status = "fail"
	}
	r.Checks = append(r.Checks, publishCheck{Status: status, Area: area, Message: message, Details: details})
}

func renderPublishReport(report *publishReport) func(t table.Writer) {
	return func(t table.Writer) {
		t.AppendHeader(table.Row{"Status", "Area", "Message", "Details"})
		for _, check := range report.Checks {
			t.AppendRow(table.Row{strings.ToUpper(check.Status), check.Area, check.Message, strings.Join(check.Details, "\n")})
		}
		published := "no"
		if report.Published {
			published = "yes"
		}
		t.AppendFooter(table.Row{strings.ToUpper(report.Status), "publish", report.OfferingID, "published: " + published})
	}
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
