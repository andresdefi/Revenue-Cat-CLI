package setup

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

type productOptions struct {
	AppID           string `json:"app_id"`
	StoreID         string `json:"store_id"`
	Type            string `json:"type"`
	DisplayName     string `json:"display_name,omitempty"`
	EntitlementKey  string `json:"entitlement_key"`
	EntitlementName string `json:"entitlement_name"`
	OfferingKey     string `json:"offering_key"`
	OfferingName    string `json:"offering_name"`
	PackageKey      string `json:"package_key"`
	PackageName     string `json:"package_name"`
	MakeCurrent     bool   `json:"make_current"`
}

type productSetupReport struct {
	Object       string          `json:"object"`
	ProjectID    string          `json:"project_id"`
	DryRun       bool            `json:"dry_run"`
	Status       string          `json:"status"`
	Options      productOptions  `json:"options"`
	Product      resourceSummary `json:"product"`
	Entitlement  resourceSummary `json:"entitlement"`
	Offering     resourceSummary `json:"offering"`
	Package      resourceSummary `json:"package"`
	Actions      []setupAction   `json:"actions"`
	NextCommands []string        `json:"next_commands"`
}

type resourceSummary struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	State string `json:"state,omitempty"`
}

type setupAction struct {
	Status  string `json:"status"`
	Area    string `json:"area"`
	Action  string `json:"action"`
	Message string `json:"message"`
}

func NewSetupCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:   "setup",
		Short: "Run guided setup workflows",
		Long: `Run guided setup workflows.

Setup commands compose lower-level RevenueCat API operations into repeatable
project configuration flows.`,
	}
	root.AddCommand(newProductCmd(projectID, outputFormat))
	return root
}

func newProductCmd(projectID, outputFormat *string) *cobra.Command {
	opts := productOptions{
		Type:            "subscription",
		EntitlementKey:  "premium",
		EntitlementName: "Premium",
		OfferingKey:     "default",
		OfferingName:    "Default",
		PackageKey:      "$rc_monthly",
		PackageName:     "Monthly",
	}

	cmd := &cobra.Command{
		Use:   "product",
		Short: "Set up a product access path",
		Long: `Set up a product access path.

This workflow creates or reuses a product, entitlement, offering, and package,
then ensures the product is attached to both the entitlement and package. It is
safe to rerun because existing resources are reused by store ID, lookup key, and
package key.`,
		Example: `  # Set up a monthly subscription path
  rc setup product \
    --app-id app1a2b3c4 \
    --store-id com.example.app.monthly \
    --display-name "Monthly" \
    --entitlement-key premium \
    --offering-key default \
    --package-key '$rc_monthly'

  # Also make the offering current
  rc setup product --app-id app1a2b3c4 --store-id com.example.app.monthly --make-current`,
		RunE: func(c *cobra.Command, args []string) error {
			if opts.DisplayName == "" {
				opts.DisplayName = opts.StoreID
			}
			if opts.EntitlementName == "" {
				opts.EntitlementName = opts.EntitlementKey
			}
			if opts.OfferingName == "" {
				opts.OfferingName = opts.OfferingKey
			}
			if opts.PackageName == "" {
				opts.PackageName = opts.PackageKey
			}

			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			report, err := runProductSetup(client, pid, opts)
			if err != nil {
				return err
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, report, renderProductSetupReport(report))
			if !api.DryRun {
				output.Success("Product setup complete")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.AppID, "app-id", "", "RevenueCat app ID (required)")
	cmd.Flags().StringVar(&opts.StoreID, "store-id", "", "store product identifier (required)")
	cmd.Flags().StringVar(&opts.Type, "type", opts.Type, "product type: subscription, one_time, consumable, non_consumable")
	cmd.Flags().StringVar(&opts.DisplayName, "display-name", "", "product display name (defaults to --store-id)")
	cmd.Flags().StringVar(&opts.EntitlementKey, "entitlement-key", opts.EntitlementKey, "entitlement lookup key")
	cmd.Flags().StringVar(&opts.EntitlementName, "entitlement-name", opts.EntitlementName, "entitlement display name")
	cmd.Flags().StringVar(&opts.OfferingKey, "offering-key", opts.OfferingKey, "offering lookup key")
	cmd.Flags().StringVar(&opts.OfferingName, "offering-name", opts.OfferingName, "offering display name")
	cmd.Flags().StringVar(&opts.PackageKey, "package-key", opts.PackageKey, "package lookup key")
	cmd.Flags().StringVar(&opts.PackageName, "package-name", opts.PackageName, "package display name")
	cmd.Flags().BoolVar(&opts.MakeCurrent, "make-current", false, "make the offering current after setup")
	cmdutil.MustMarkFlagRequired(cmd, "app-id")
	cmdutil.MustMarkFlagRequired(cmd, "store-id")
	return cmd
}

func runProductSetup(client *api.Client, projectID string, opts productOptions) (*productSetupReport, error) {
	report := &productSetupReport{
		Object:    "product_setup",
		ProjectID: projectID,
		DryRun:    api.DryRun,
		Status:    "ok",
		Options:   opts,
	}

	product, productExists, err := ensureProduct(client, projectID, opts, report)
	if err != nil {
		return nil, err
	}
	entitlement, entitlementExists, err := ensureEntitlement(client, projectID, opts, report)
	if err != nil {
		return nil, err
	}
	offering, offeringExists, err := ensureOffering(client, projectID, opts, report)
	if err != nil {
		return nil, err
	}
	pkg, packageExists, err := ensurePackage(client, projectID, offering.ID, opts, report)
	if err != nil {
		return nil, err
	}

	if err := ensureEntitlementProduct(client, projectID, entitlement.ID, product.ID, entitlementExists && productExists, report); err != nil {
		return nil, err
	}
	if err := ensurePackageProduct(client, projectID, pkg.ID, product.ID, packageExists && productExists, report); err != nil {
		return nil, err
	}
	if opts.MakeCurrent {
		if err := ensureOfferingCurrent(client, projectID, offering, offeringExists, report); err != nil {
			return nil, err
		}
	}

	report.Product = resourceSummary{ID: product.ID, Label: opts.StoreID, State: product.State}
	report.Entitlement = resourceSummary{ID: entitlement.ID, Label: opts.EntitlementKey, State: entitlement.State}
	report.Offering = resourceSummary{ID: offering.ID, Label: opts.OfferingKey, State: offering.State}
	report.Package = resourceSummary{ID: pkg.ID, Label: opts.PackageKey}
	report.NextCommands = []string{
		fmt.Sprintf("rc products get %s", product.ID),
		fmt.Sprintf("rc entitlements products %s --all", entitlement.ID),
		fmt.Sprintf("rc packages products %s --all", pkg.ID),
		fmt.Sprintf("rc offerings publish %s", offering.ID),
	}
	return report, nil
}

func ensureProduct(client *api.Client, projectID string, opts productOptions, report *productSetupReport) (api.Product, bool, error) {
	products, err := api.PaginateAll[api.Product](client, fmt.Sprintf("/projects/%s/products", url.PathEscape(projectID)), nil)
	if err != nil {
		return api.Product{}, false, fmt.Errorf("list products: %w", err)
	}
	for _, product := range products {
		if product.AppID == opts.AppID && product.StoreIdentifier == opts.StoreID && product.Type == opts.Type {
			report.add("reused", "product", "reuse", fmt.Sprintf("Reused product %s for %s.", product.ID, opts.StoreID))
			return product, true, nil
		}
	}
	if api.DryRun {
		report.add("planned", "product", "create", fmt.Sprintf("Would create product %s.", opts.StoreID))
		return api.Product{ID: "(new product)", StoreIdentifier: opts.StoreID, Type: opts.Type, State: "planned", AppID: opts.AppID}, false, nil
	}
	body := map[string]any{"store_identifier": opts.StoreID, "app_id": opts.AppID, "type": opts.Type}
	if opts.DisplayName != "" {
		body["display_name"] = opts.DisplayName
	}
	data, err := client.Post(fmt.Sprintf("/projects/%s/products", url.PathEscape(projectID)), body)
	if err != nil {
		return api.Product{}, false, fmt.Errorf("create product: %w", err)
	}
	var product api.Product
	if err := json.Unmarshal(data, &product); err != nil {
		return api.Product{}, false, fmt.Errorf("parse created product: %w", err)
	}
	report.add("changed", "product", "create", fmt.Sprintf("Created product %s for %s.", product.ID, opts.StoreID))
	return product, false, nil
}

func ensureEntitlement(client *api.Client, projectID string, opts productOptions, report *productSetupReport) (api.Entitlement, bool, error) {
	entitlements, err := api.PaginateAll[api.Entitlement](client, fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(projectID)), nil)
	if err != nil {
		return api.Entitlement{}, false, fmt.Errorf("list entitlements: %w", err)
	}
	for _, entitlement := range entitlements {
		if entitlement.LookupKey == opts.EntitlementKey {
			report.add("reused", "entitlement", "reuse", fmt.Sprintf("Reused entitlement %s.", entitlement.ID))
			return entitlement, true, nil
		}
	}
	if api.DryRun {
		report.add("planned", "entitlement", "create", fmt.Sprintf("Would create entitlement %s.", opts.EntitlementKey))
		return api.Entitlement{ID: "(new entitlement)", LookupKey: opts.EntitlementKey, DisplayName: opts.EntitlementName, State: "planned"}, false, nil
	}
	data, err := client.Post(fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(projectID)), map[string]any{"lookup_key": opts.EntitlementKey, "display_name": opts.EntitlementName})
	if err != nil {
		return api.Entitlement{}, false, fmt.Errorf("create entitlement: %w", err)
	}
	var entitlement api.Entitlement
	if err := json.Unmarshal(data, &entitlement); err != nil {
		return api.Entitlement{}, false, fmt.Errorf("parse created entitlement: %w", err)
	}
	report.add("changed", "entitlement", "create", fmt.Sprintf("Created entitlement %s.", entitlement.ID))
	return entitlement, false, nil
}

func ensureOffering(client *api.Client, projectID string, opts productOptions, report *productSetupReport) (api.Offering, bool, error) {
	offerings, err := api.PaginateAll[api.Offering](client, fmt.Sprintf("/projects/%s/offerings", url.PathEscape(projectID)), nil)
	if err != nil {
		return api.Offering{}, false, fmt.Errorf("list offerings: %w", err)
	}
	for _, offering := range offerings {
		if offering.LookupKey == opts.OfferingKey {
			report.add("reused", "offering", "reuse", fmt.Sprintf("Reused offering %s.", offering.ID))
			return offering, true, nil
		}
	}
	if api.DryRun {
		report.add("planned", "offering", "create", fmt.Sprintf("Would create offering %s.", opts.OfferingKey))
		return api.Offering{ID: "(new offering)", LookupKey: opts.OfferingKey, DisplayName: opts.OfferingName, State: "planned"}, false, nil
	}
	data, err := client.Post(fmt.Sprintf("/projects/%s/offerings", url.PathEscape(projectID)), map[string]any{"lookup_key": opts.OfferingKey, "display_name": opts.OfferingName})
	if err != nil {
		return api.Offering{}, false, fmt.Errorf("create offering: %w", err)
	}
	var offering api.Offering
	if err := json.Unmarshal(data, &offering); err != nil {
		return api.Offering{}, false, fmt.Errorf("parse created offering: %w", err)
	}
	report.add("changed", "offering", "create", fmt.Sprintf("Created offering %s.", offering.ID))
	return offering, false, nil
}

func ensurePackage(client *api.Client, projectID, offeringID string, opts productOptions, report *productSetupReport) (api.Package, bool, error) {
	if api.DryRun && offeringID == "(new offering)" {
		report.add("planned", "package", "create", fmt.Sprintf("Would create package %s.", opts.PackageKey))
		return api.Package{ID: "(new package)", LookupKey: opts.PackageKey, DisplayName: opts.PackageName}, false, nil
	}
	packages, err := api.PaginateAll[api.Package](client, fmt.Sprintf("/projects/%s/offerings/%s/packages", url.PathEscape(projectID), url.PathEscape(offeringID)), nil)
	if err != nil {
		return api.Package{}, false, fmt.Errorf("list packages: %w", err)
	}
	for _, pkg := range packages {
		if pkg.LookupKey == opts.PackageKey {
			report.add("reused", "package", "reuse", fmt.Sprintf("Reused package %s.", pkg.ID))
			return pkg, true, nil
		}
	}
	if api.DryRun {
		report.add("planned", "package", "create", fmt.Sprintf("Would create package %s.", opts.PackageKey))
		return api.Package{ID: "(new package)", LookupKey: opts.PackageKey, DisplayName: opts.PackageName}, false, nil
	}
	data, err := client.Post(fmt.Sprintf("/projects/%s/offerings/%s/packages", url.PathEscape(projectID), url.PathEscape(offeringID)), map[string]any{"lookup_key": opts.PackageKey, "display_name": opts.PackageName})
	if err != nil {
		return api.Package{}, false, fmt.Errorf("create package: %w", err)
	}
	var pkg api.Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		return api.Package{}, false, fmt.Errorf("parse created package: %w", err)
	}
	report.add("changed", "package", "create", fmt.Sprintf("Created package %s.", pkg.ID))
	return pkg, false, nil
}

func ensureEntitlementProduct(client *api.Client, projectID, entitlementID, productID string, canCheck bool, report *productSetupReport) error {
	if api.DryRun {
		report.add("planned", "entitlement", "attach-product", fmt.Sprintf("Would attach product %s to entitlement %s.", productID, entitlementID))
		return nil
	}
	if canCheck {
		products, err := api.PaginateAll[api.Product](client, fmt.Sprintf("/projects/%s/entitlements/%s/products", url.PathEscape(projectID), url.PathEscape(entitlementID)), nil)
		if err != nil {
			return fmt.Errorf("list entitlement products: %w", err)
		}
		for _, product := range products {
			if product.ID == productID {
				report.add("reused", "entitlement", "attach-product", "Product already attached to entitlement.")
				return nil
			}
		}
	}
	_, err := client.Post(fmt.Sprintf("/projects/%s/entitlements/%s/actions/attach_products", url.PathEscape(projectID), url.PathEscape(entitlementID)), map[string]any{"product_ids": []string{productID}})
	if err != nil {
		return fmt.Errorf("attach product to entitlement: %w", err)
	}
	report.add("changed", "entitlement", "attach-product", fmt.Sprintf("Attached product %s to entitlement %s.", productID, entitlementID))
	return nil
}

func ensurePackageProduct(client *api.Client, projectID, packageID, productID string, canCheck bool, report *productSetupReport) error {
	if api.DryRun {
		report.add("planned", "package", "attach-product", fmt.Sprintf("Would attach product %s to package %s.", productID, packageID))
		return nil
	}
	if canCheck {
		products, err := api.PaginateAll[api.PackageProduct](client, fmt.Sprintf("/projects/%s/packages/%s/products", url.PathEscape(projectID), url.PathEscape(packageID)), nil)
		if err != nil {
			return fmt.Errorf("list package products: %w", err)
		}
		for _, product := range products {
			if product.ProductID == productID {
				report.add("reused", "package", "attach-product", "Product already attached to package.")
				return nil
			}
		}
	}
	_, err := client.Post(fmt.Sprintf("/projects/%s/packages/%s/actions/attach_products", url.PathEscape(projectID), url.PathEscape(packageID)), map[string]any{"products": []map[string]any{{"product_id": productID, "eligibility_criteria": "all"}}})
	if err != nil {
		return fmt.Errorf("attach product to package: %w", err)
	}
	report.add("changed", "package", "attach-product", fmt.Sprintf("Attached product %s to package %s.", productID, packageID))
	return nil
}

func ensureOfferingCurrent(client *api.Client, projectID string, offering api.Offering, offeringExists bool, report *productSetupReport) error {
	if offeringExists && offering.IsCurrent {
		report.add("reused", "offering", "publish", "Offering is already current.")
		return nil
	}
	if api.DryRun {
		report.add("planned", "offering", "publish", fmt.Sprintf("Would make offering %s current.", offering.ID))
		return nil
	}
	_, err := client.Post(fmt.Sprintf("/projects/%s/offerings/%s", url.PathEscape(projectID), url.PathEscape(offering.ID)), map[string]any{"is_current": true})
	if err != nil {
		return fmt.Errorf("make offering current: %w", err)
	}
	report.add("changed", "offering", "publish", fmt.Sprintf("Made offering %s current.", offering.ID))
	return nil
}

func (r *productSetupReport) add(status, area, action, message string) {
	if status == "changed" {
		r.Status = "changed"
	}
	if status == "planned" && r.Status != "changed" {
		r.Status = "planned"
	}
	r.Actions = append(r.Actions, setupAction{Status: status, Area: area, Action: action, Message: message})
}

func renderProductSetupReport(report *productSetupReport) func(t table.Writer) {
	return func(t table.Writer) {
		t.AppendHeader(table.Row{"Status", "Area", "Action", "Message"})
		for _, action := range report.Actions {
			t.AppendRow(table.Row{action.Status, action.Area, action.Action, action.Message})
		}
		t.AppendSeparator()
		t.AppendRows([]table.Row{
			{"resource", "product", report.Product.ID, report.Product.Label},
			{"resource", "entitlement", report.Entitlement.ID, report.Entitlement.Label},
			{"resource", "offering", report.Offering.ID, report.Offering.Label},
			{"resource", "package", report.Package.ID, report.Package.Label},
		})
		t.AppendFooter(table.Row{report.Status, "setup", "product", fmt.Sprintf("%d action(s)", len(report.Actions))})
	}
}
