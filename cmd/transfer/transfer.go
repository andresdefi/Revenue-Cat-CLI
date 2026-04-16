package transfer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// ProjectConfig is the top-level export format for a project configuration.
type ProjectConfig struct {
	Version      string              `json:"version"`
	ExportedAt   string              `json:"exported_at"`
	Apps         []api.App           `json:"apps,omitempty"`
	Products     []api.Product       `json:"products"`
	Entitlements []EntitlementConfig `json:"entitlements"`
	Offerings    []OfferingConfig    `json:"offerings"`
}

type EntitlementConfig struct {
	Object      string   `json:"object,omitempty"`
	ID          string   `json:"id"`
	ProjectID   string   `json:"project_id,omitempty"`
	LookupKey   string   `json:"lookup_key"`
	DisplayName string   `json:"display_name"`
	State       string   `json:"state,omitempty"`
	CreatedAt   int64    `json:"created_at,omitempty"`
	ProductIDs  []string `json:"product_ids,omitempty"`
}

type OfferingConfig struct {
	Object      string          `json:"object,omitempty"`
	ID          string          `json:"id"`
	ProjectID   string          `json:"project_id,omitempty"`
	LookupKey   string          `json:"lookup_key"`
	DisplayName string          `json:"display_name"`
	IsCurrent   bool            `json:"is_current,omitempty"`
	State       string          `json:"state,omitempty"`
	CreatedAt   int64           `json:"created_at,omitempty"`
	Metadata    map[string]any  `json:"metadata,omitempty"`
	Packages    []PackageConfig `json:"packages,omitempty"`
}

type PackageConfig struct {
	Object      string                 `json:"object,omitempty"`
	ID          string                 `json:"id"`
	LookupKey   string                 `json:"lookup_key"`
	DisplayName string                 `json:"display_name"`
	Position    *int                   `json:"position,omitempty"`
	CreatedAt   int64                  `json:"created_at,omitempty"`
	Products    []PackageProductConfig `json:"products,omitempty"`
}

type PackageProductConfig struct {
	ProductID           string `json:"product_id"`
	EligibilityCriteria string `json:"eligibility_criteria"`
}

type MigrationPlan struct {
	Object          string            `json:"object"`
	SourceProjectID string            `json:"source_project_id"`
	TargetProjectID string            `json:"target_project_id"`
	DryRun          bool              `json:"dry_run"`
	Status          string            `json:"status"`
	Counts          MigrationCounts   `json:"counts"`
	Warnings        []string          `json:"warnings,omitempty"`
	Actions         []MigrationAction `json:"actions"`
}

type MigrationCounts struct {
	Products        int `json:"products"`
	Entitlements    int `json:"entitlements"`
	Offerings       int `json:"offerings"`
	Packages        int `json:"packages"`
	PackageProducts int `json:"package_products"`
	Create          int `json:"create"`
	Reuse           int `json:"reuse"`
	Skip            int `json:"skip"`
	Update          int `json:"update"`
	Archive         int `json:"archive"`
}

type MigrationAction struct {
	Status   string `json:"status"`
	Area     string `json:"area"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
	Message  string `json:"message"`
}

// NewExportCmd creates the `rc export` command.
func NewExportCmd(projectID, outputFormat *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export project configuration",
		Long: `Export project configuration as a JSON file.
This includes apps, products, entitlements and their product attachments,
offerings, packages, package-product attachments, metadata, and current/archive state.`,
		Example: `  # Export project config
  rc export --file config.json

  # Export from a specific project
  rc export --file config.json --project proj1a2b3c4d5`,
		RunE: func(c *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}

			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			config, err := exportProjectConfig(client, pid)
			if err != nil {
				return err
			}

			data, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			if err := os.WriteFile(file, data, 0o600); err != nil { //nolint:gosec // user-specified export path
				return fmt.Errorf("failed to write file: %w", err)
			}

			output.Success("Exported %d apps, %d products, %d entitlements, %d offerings to %s",
				len(config.Apps), len(config.Products), len(config.Entitlements), len(config.Offerings), file)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "output file path (required)")
	cmdutil.MarkBeta(cmd)
	return cmd
}

// NewImportCmd creates the `rc import` command.
func NewImportCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		file    string
		appMaps []string
	)

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import project configuration from a JSON file",
		Long: `Import a project configuration exported with 'rc export'.
Creates or reuses matching products, entitlements, offerings, and packages,
then restores product attachments, package attachments, metadata, and state.

Apps are matched to existing target apps by ID or by name and type. Use
--app-map source_app_id=target_app_id when automatic matching is not enough.`,
		Example: `  # Import into the default project
  rc import --file config.json

  # Import into a specific project
  rc import --file config.json --project proj_target123

  # Map a source app ID to a target app ID
  rc import --file config.json --app-map app_source=app_target`,
		RunE: func(c *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}

			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var config ProjectConfig
			if err := json.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("failed to parse config file: %w", err)
			}

			result, err := importProjectConfig(client, pid, config, appMaps)
			if err != nil {
				return err
			}

			output.Success("Imported %d/%d products, %d/%d entitlements, %d/%d offerings, %d/%d packages",
				result.Products, len(config.Products),
				result.Entitlements, len(config.Entitlements),
				result.Offerings, len(config.Offerings),
				result.Packages, countPackages(config.Offerings))
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "input file path (required)")
	cmd.Flags().StringArrayVar(&appMaps, "app-map", nil, "map source app ID to target app ID (source=target, repeatable)")
	cmdutil.MarkBeta(cmd)
	return cmd
}

func NewMigrateCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:   "migrate",
		Short: "Plan project migration workflows",
		Long: `Plan project migration workflows.

Migration commands provide a safer UX around export/import behavior. The
project workflow currently supports dry-run planning only and never mutates the
target project.`,
	}
	root.AddCommand(newMigrateProjectCmd(projectID, outputFormat))
	cmdutil.MarkBeta(root)
	return root
}

func newMigrateProjectCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		sourceProjectID string
		targetProjectID string
		appMaps         []string
	)
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Dry-run a project configuration migration",
		Long: `Dry-run a project configuration migration.

The command exports the source project in memory, compares it with the target
project, and reports what import would create, reuse, update, attach, archive,
or skip. It requires --dry-run so migration planning stays read-only.`,
		Example: `  # Plan a migration into the default project
  rc migrate project --source-project proj_source --dry-run

  # Plan a migration into an explicit target project
  rc migrate project --source-project proj_source --target-project proj_target --dry-run

  # Include explicit app ID mappings
  rc migrate project --source-project proj_source --target-project proj_target --app-map app_source=app_target --dry-run`,
		RunE: func(c *cobra.Command, args []string) error {
			if !api.DryRun {
				return fmt.Errorf("migrate project requires --dry-run")
			}
			if sourceProjectID == "" {
				return fmt.Errorf("--source-project is required")
			}
			targetID := targetProjectID
			if targetID == "" {
				var err error
				targetID, err = cmdutil.ResolveProject(projectID)
				if err != nil {
					return err
				}
			}
			if sourceProjectID == targetID {
				return fmt.Errorf("source and target projects must be different")
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			plan, err := dryRunProjectMigration(client, sourceProjectID, targetID, appMaps)
			if err != nil {
				return err
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, plan, renderMigrationPlan(plan))
			return nil
		},
	}
	cmd.Flags().StringVar(&sourceProjectID, "source-project", "", "source project ID (required)")
	cmd.Flags().StringVar(&targetProjectID, "target-project", "", "target project ID (defaults to --project/default project)")
	cmd.Flags().StringArrayVar(&appMaps, "app-map", nil, "map source app ID to target app ID (source=target, repeatable)")
	return cmd
}

func dryRunProjectMigration(client *api.Client, sourceProjectID, targetProjectID string, appMaps []string) (*MigrationPlan, error) {
	config, err := exportProjectConfig(client, sourceProjectID)
	if err != nil {
		return nil, err
	}
	targetApps, err := api.PaginateAll[api.App](client, fmt.Sprintf("/projects/%s/apps", url.PathEscape(targetProjectID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch target apps: %w", err)
	}
	targetProducts, err := api.PaginateAll[api.Product](client, fmt.Sprintf("/projects/%s/products", url.PathEscape(targetProjectID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch target products: %w", err)
	}
	targetEntitlements, err := api.PaginateAll[api.Entitlement](client, fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(targetProjectID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch target entitlements: %w", err)
	}
	targetOfferings, err := api.PaginateAll[api.Offering](client, fmt.Sprintf("/projects/%s/offerings", url.PathEscape(targetProjectID)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch target offerings: %w", err)
	}

	appMap, err := buildAppMap(config.Apps, targetApps, appMaps)
	if err != nil {
		return nil, err
	}
	plan := &MigrationPlan{
		Object:          "project_migration_plan",
		SourceProjectID: sourceProjectID,
		TargetProjectID: targetProjectID,
		DryRun:          true,
		Status:          "ok",
	}
	productMap := planProducts(plan, config.Products, targetProducts, appMap)
	entitlementMap := planEntitlements(plan, config.Entitlements, targetEntitlements, productMap)
	offeringMap, packageMap := planOfferings(client, plan, targetProjectID, config.Offerings, targetOfferings, productMap)
	planArchives(plan, config.Products, config.Entitlements, config.Offerings, productMap, entitlementMap, offeringMap, packageMap)
	if len(plan.Warnings) > 0 || plan.Counts.Skip > 0 {
		plan.Status = "warn"
	}
	return plan, nil
}

func planProducts(plan *MigrationPlan, products, targetProducts []api.Product, appMap map[string]string) map[string]string {
	productMap := make(map[string]string, len(products))
	existing := make(map[string]api.Product, len(targetProducts))
	for _, product := range targetProducts {
		existing[productKey(product.AppID, product.StoreIdentifier, product.Type)] = product
	}
	plan.Counts.Products = len(products)
	for _, product := range products {
		targetAppID := appMap[product.AppID]
		if targetAppID == "" {
			plan.warn(fmt.Sprintf("No target app mapping for product %s from source app %s.", product.StoreIdentifier, product.AppID))
			plan.add("skip", "products", "skip", product.StoreIdentifier, "No target app mapping.")
			continue
		}
		if found, ok := existing[productKey(targetAppID, product.StoreIdentifier, product.Type)]; ok {
			productMap[product.ID] = found.ID
			plan.add("reuse", "products", "reuse", product.StoreIdentifier, fmt.Sprintf("Would reuse target product %s.", found.ID))
			continue
		}
		productMap[product.ID] = "(new product)"
		plan.add("create", "products", "create", product.StoreIdentifier, "Would create product.")
	}
	return productMap
}

func planEntitlements(plan *MigrationPlan, entitlements []EntitlementConfig, targetEntitlements []api.Entitlement, productMap map[string]string) map[string]string {
	entitlementMap := make(map[string]string, len(entitlements))
	existing := make(map[string]api.Entitlement, len(targetEntitlements))
	for _, entitlement := range targetEntitlements {
		existing[entitlement.LookupKey] = entitlement
	}
	plan.Counts.Entitlements = len(entitlements)
	for _, entitlement := range entitlements {
		targetID := "(new entitlement)"
		if found, ok := existing[entitlement.LookupKey]; ok {
			targetID = found.ID
			plan.add("reuse", "entitlements", "reuse", entitlement.LookupKey, fmt.Sprintf("Would reuse target entitlement %s.", found.ID))
			if found.DisplayName != entitlement.DisplayName {
				plan.add("update", "entitlements", "update", entitlement.LookupKey, "Would update display name.")
			}
		} else {
			plan.add("create", "entitlements", "create", entitlement.LookupKey, "Would create entitlement.")
		}
		entitlementMap[entitlement.ID] = targetID
		for _, sourceProductID := range entitlement.ProductIDs {
			if productMap[sourceProductID] == "" {
				plan.add("skip", "entitlements", "attach-product", entitlement.LookupKey, fmt.Sprintf("Would skip missing product attachment %s.", sourceProductID))
				continue
			}
			plan.add("update", "entitlements", "attach-product", entitlement.LookupKey, fmt.Sprintf("Would attach mapped product for %s.", sourceProductID))
		}
	}
	return entitlementMap
}

func planOfferings(client *api.Client, plan *MigrationPlan, targetProjectID string, offerings []OfferingConfig, targetOfferings []api.Offering, productMap map[string]string) (map[string]string, map[string]string) {
	offeringMap := make(map[string]string, len(offerings))
	packageMap := map[string]string{}
	existing := make(map[string]api.Offering, len(targetOfferings))
	for _, offering := range targetOfferings {
		existing[offering.LookupKey] = offering
	}
	plan.Counts.Offerings = len(offerings)
	for _, offering := range offerings {
		targetID := "(new offering)"
		targetPackages := []api.Package{}
		if found, ok := existing[offering.LookupKey]; ok {
			targetID = found.ID
			plan.add("reuse", "offerings", "reuse", offering.LookupKey, fmt.Sprintf("Would reuse target offering %s.", found.ID))
			if found.DisplayName != offering.DisplayName || found.IsCurrent != offering.IsCurrent {
				plan.add("update", "offerings", "update", offering.LookupKey, "Would update offering fields.")
			}
			var err error
			targetPackages, err = api.PaginateAll[api.Package](client, fmt.Sprintf("/projects/%s/offerings/%s/packages", url.PathEscape(targetProjectID), url.PathEscape(found.ID)), nil)
			if err != nil {
				plan.warn(fmt.Sprintf("Could not inspect target packages for offering %s: %v", offering.LookupKey, err))
			}
		} else {
			plan.add("create", "offerings", "create", offering.LookupKey, "Would create offering.")
		}
		offeringMap[offering.ID] = targetID
		planPackages(plan, offering, targetPackages, productMap, packageMap)
	}
	return offeringMap, packageMap
}

func planPackages(plan *MigrationPlan, offering OfferingConfig, targetPackages []api.Package, productMap map[string]string, packageMap map[string]string) {
	existing := make(map[string]api.Package, len(targetPackages))
	for _, pkg := range targetPackages {
		existing[pkg.LookupKey] = pkg
	}
	plan.Counts.Packages += len(offering.Packages)
	for _, pkg := range offering.Packages {
		targetID := "(new package)"
		if found, ok := existing[pkg.LookupKey]; ok {
			targetID = found.ID
			plan.add("reuse", "packages", "reuse", pkg.LookupKey, fmt.Sprintf("Would reuse target package %s.", found.ID))
			if found.DisplayName != pkg.DisplayName {
				plan.add("update", "packages", "update", pkg.LookupKey, "Would update display name.")
			}
		} else {
			plan.add("create", "packages", "create", pkg.LookupKey, "Would create package.")
		}
		packageMap[pkg.ID] = targetID
		plan.Counts.PackageProducts += len(pkg.Products)
		for _, product := range pkg.Products {
			if productMap[product.ProductID] == "" {
				plan.add("skip", "packages", "attach-product", pkg.LookupKey, fmt.Sprintf("Would skip missing product attachment %s.", product.ProductID))
				continue
			}
			plan.add("update", "packages", "attach-product", pkg.LookupKey, fmt.Sprintf("Would attach mapped product for %s.", product.ProductID))
		}
	}
}

func planArchives(plan *MigrationPlan, products []api.Product, entitlements []EntitlementConfig, offerings []OfferingConfig, productMap, entitlementMap, offeringMap, packageMap map[string]string) {
	for _, product := range products {
		if product.State == "archived" && productMap[product.ID] != "" {
			plan.add("archive", "products", "archive", product.StoreIdentifier, "Would archive target product.")
		}
	}
	for _, entitlement := range entitlements {
		if entitlement.State == "archived" && entitlementMap[entitlement.ID] != "" {
			plan.add("archive", "entitlements", "archive", entitlement.LookupKey, "Would archive target entitlement.")
		}
	}
	for _, offering := range offerings {
		if offering.State == "archived" && offeringMap[offering.ID] != "" {
			plan.add("archive", "offerings", "archive", offering.LookupKey, "Would archive target offering.")
		}
		for _, pkg := range offering.Packages {
			if packageMap[pkg.ID] == "" {
				continue
			}
		}
	}
}

func (p *MigrationPlan) warn(message string) {
	p.Warnings = append(p.Warnings, message)
}

func (p *MigrationPlan) add(status, area, action, resource, message string) {
	switch status {
	case "create":
		p.Counts.Create++
	case "reuse":
		p.Counts.Reuse++
	case "skip":
		p.Counts.Skip++
	case "update":
		p.Counts.Update++
	case "archive":
		p.Counts.Archive++
	}
	p.Actions = append(p.Actions, MigrationAction{Status: status, Area: area, Action: action, Resource: resource, Message: message})
}

func renderMigrationPlan(plan *MigrationPlan) func(t table.Writer) {
	return func(t table.Writer) {
		t.AppendHeader(table.Row{"Status", "Area", "Action", "Resource", "Message"})
		for _, action := range plan.Actions {
			t.AppendRow(table.Row{action.Status, action.Area, action.Action, action.Resource, action.Message})
		}
		if len(plan.Warnings) > 0 {
			t.AppendSeparator()
			for _, warning := range plan.Warnings {
				t.AppendRow(table.Row{"warn", "migration", "warning", "", warning})
			}
		}
		t.AppendFooter(table.Row{
			plan.Status,
			"migration",
			"dry-run",
			fmt.Sprintf("%s -> %s", plan.SourceProjectID, plan.TargetProjectID),
			fmt.Sprintf("create=%d reuse=%d update=%d skip=%d archive=%d", plan.Counts.Create, plan.Counts.Reuse, plan.Counts.Update, plan.Counts.Skip, plan.Counts.Archive),
		})
	}
}

func exportProjectConfig(client *api.Client, pid string) (*ProjectConfig, error) {
	apps, err := api.PaginateAll[api.App](client, fmt.Sprintf("/projects/%s/apps", url.PathEscape(pid)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch apps: %w", err)
	}
	products, err := api.PaginateAll[api.Product](client, fmt.Sprintf("/projects/%s/products", url.PathEscape(pid)), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}
	entitlements, err := exportEntitlements(client, pid)
	if err != nil {
		return nil, err
	}
	offerings, err := exportOfferings(client, pid)
	if err != nil {
		return nil, err
	}

	return &ProjectConfig{
		Version:      "2",
		ExportedAt:   time.Now().UTC().Format(time.RFC3339),
		Apps:         apps,
		Products:     products,
		Entitlements: entitlements,
		Offerings:    offerings,
	}, nil
}

func exportEntitlements(client *api.Client, pid string) ([]EntitlementConfig, error) {
	path := fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(pid))
	items, err := api.PaginateAll[api.Entitlement](client, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch entitlements: %w", err)
	}

	out := make([]EntitlementConfig, 0, len(items))
	for _, ent := range items {
		productsPath := fmt.Sprintf("/projects/%s/entitlements/%s/products", url.PathEscape(pid), url.PathEscape(ent.ID))
		products, err := api.PaginateAll[api.Product](client, productsPath, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch products for entitlement %s: %w", ent.ID, err)
		}
		productIDs := make([]string, 0, len(products))
		for _, product := range products {
			productIDs = append(productIDs, product.ID)
		}
		out = append(out, EntitlementConfig{
			Object:      ent.Object,
			ID:          ent.ID,
			ProjectID:   ent.ProjectID,
			LookupKey:   ent.LookupKey,
			DisplayName: ent.DisplayName,
			State:       ent.State,
			CreatedAt:   ent.CreatedAt,
			ProductIDs:  productIDs,
		})
	}
	return out, nil
}

func exportOfferings(client *api.Client, pid string) ([]OfferingConfig, error) {
	path := fmt.Sprintf("/projects/%s/offerings", url.PathEscape(pid))
	items, err := api.PaginateAll[api.Offering](client, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch offerings: %w", err)
	}

	out := make([]OfferingConfig, 0, len(items))
	for _, offering := range items {
		packagesPath := fmt.Sprintf("/projects/%s/offerings/%s/packages", url.PathEscape(pid), url.PathEscape(offering.ID))
		packages, err := api.PaginateAll[api.Package](client, packagesPath, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch packages for offering %s: %w", offering.ID, err)
		}
		packageConfigs := make([]PackageConfig, 0, len(packages))
		for _, pkg := range packages {
			productsPath := fmt.Sprintf("/projects/%s/packages/%s/products", url.PathEscape(pid), url.PathEscape(pkg.ID))
			products, err := api.PaginateAll[api.PackageProduct](client, productsPath, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch products for package %s: %w", pkg.ID, err)
			}
			packageConfigs = append(packageConfigs, packageConfig(pkg, products))
		}
		out = append(out, OfferingConfig{
			Object:      offering.Object,
			ID:          offering.ID,
			ProjectID:   offering.ProjectID,
			LookupKey:   offering.LookupKey,
			DisplayName: offering.DisplayName,
			IsCurrent:   offering.IsCurrent,
			State:       offering.State,
			CreatedAt:   offering.CreatedAt,
			Metadata:    offering.Metadata,
			Packages:    packageConfigs,
		})
	}
	return out, nil
}

func packageConfig(pkg api.Package, products []api.PackageProduct) PackageConfig {
	packageProducts := make([]PackageProductConfig, 0, len(products))
	for _, product := range products {
		packageProducts = append(packageProducts, PackageProductConfig{
			ProductID:           product.ProductID,
			EligibilityCriteria: product.EligibilityCriteria,
		})
	}
	return PackageConfig{
		Object:      pkg.Object,
		ID:          pkg.ID,
		LookupKey:   pkg.LookupKey,
		DisplayName: pkg.DisplayName,
		Position:    pkg.Position,
		CreatedAt:   pkg.CreatedAt,
		Products:    packageProducts,
	}
}

type importResult struct {
	Products     int
	Entitlements int
	Offerings    int
	Packages     int
}

func importProjectConfig(client *api.Client, pid string, config ProjectConfig, appMaps []string) (importResult, error) {
	result := importResult{}
	targetApps, err := api.PaginateAll[api.App](client, fmt.Sprintf("/projects/%s/apps", url.PathEscape(pid)), nil)
	if err != nil {
		return result, fmt.Errorf("failed to fetch target apps: %w", err)
	}
	appMap, err := buildAppMap(config.Apps, targetApps, appMaps)
	if err != nil {
		return result, err
	}

	targetProducts, err := api.PaginateAll[api.Product](client, fmt.Sprintf("/projects/%s/products", url.PathEscape(pid)), nil)
	if err != nil {
		return result, fmt.Errorf("failed to fetch target products: %w", err)
	}
	productMap := importProducts(client, pid, config.Products, targetProducts, appMap, &result)

	targetEntitlements, err := api.PaginateAll[api.Entitlement](client, fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(pid)), nil)
	if err != nil {
		return result, fmt.Errorf("failed to fetch target entitlements: %w", err)
	}
	entitlementMap := importEntitlements(client, pid, config.Entitlements, targetEntitlements, productMap, &result)

	targetOfferings, err := api.PaginateAll[api.Offering](client, fmt.Sprintf("/projects/%s/offerings", url.PathEscape(pid)), nil)
	if err != nil {
		return result, fmt.Errorf("failed to fetch target offerings: %w", err)
	}
	offeringMap := importOfferings(client, pid, config.Offerings, targetOfferings, productMap, &result)

	archiveImportedEntities(client, pid, config.Products, config.Entitlements, config.Offerings, productMap, entitlementMap, offeringMap)
	return result, nil
}

func buildAppMap(sourceApps, targetApps []api.App, mappings []string) (map[string]string, error) {
	out := make(map[string]string, len(sourceApps)+len(mappings))
	targetByID := make(map[string]api.App, len(targetApps))
	targetByNameType := make(map[string]api.App, len(targetApps))
	for _, app := range targetApps {
		targetByID[app.ID] = app
		targetByNameType[appKey(app.Name, app.Type)] = app
	}

	for _, mapping := range mappings {
		source, target, ok := strings.Cut(mapping, "=")
		if !ok || source == "" || target == "" {
			return nil, fmt.Errorf("invalid --app-map %q (expected source_app_id=target_app_id)", mapping)
		}
		if _, ok := targetByID[target]; !ok {
			output.Warn("Mapped target app %s was not found in the target project", target)
		}
		out[source] = target
	}

	for _, source := range sourceApps {
		if _, ok := out[source.ID]; ok {
			continue
		}
		if _, ok := targetByID[source.ID]; ok {
			out[source.ID] = source.ID
			continue
		}
		if target, ok := targetByNameType[appKey(source.Name, source.Type)]; ok {
			out[source.ID] = target.ID
		}
	}
	return out, nil
}

func importProducts(client *api.Client, pid string, products, targetProducts []api.Product, appMap map[string]string, result *importResult) map[string]string {
	productMap := make(map[string]string, len(products))
	existing := make(map[string]api.Product, len(targetProducts))
	for _, product := range targetProducts {
		existing[productKey(product.AppID, product.StoreIdentifier, product.Type)] = product
	}

	total := len(products)
	for i, product := range products {
		output.Progress(i+1, total, "Importing product %s", product.StoreIdentifier)
		targetAppID, ok := appMap[product.AppID]
		if !ok {
			output.Warn("Skipping product %s: no target app mapping for source app %s", product.StoreIdentifier, product.AppID)
			continue
		}

		if found, ok := existing[productKey(targetAppID, product.StoreIdentifier, product.Type)]; ok {
			productMap[product.ID] = found.ID
			result.Products++
			continue
		}

		body := map[string]any{
			"store_identifier": product.StoreIdentifier,
			"app_id":           targetAppID,
			"type":             product.Type,
		}
		if product.DisplayName != nil {
			body["display_name"] = *product.DisplayName
		}
		data, err := client.Post(fmt.Sprintf("/projects/%s/products", url.PathEscape(pid)), body)
		if err != nil {
			output.Warn("Failed to create product %s: %v", product.StoreIdentifier, err)
			continue
		}
		var created api.Product
		if err := json.Unmarshal(data, &created); err != nil {
			output.Warn("Failed to parse created product %s: %v", product.StoreIdentifier, err)
			continue
		}
		productMap[product.ID] = created.ID
		result.Products++
	}
	return productMap
}

func importEntitlements(client *api.Client, pid string, entitlements []EntitlementConfig, targetEntitlements []api.Entitlement, productMap map[string]string, result *importResult) map[string]string {
	entitlementMap := make(map[string]string, len(entitlements))
	existing := make(map[string]api.Entitlement, len(targetEntitlements))
	for _, entitlement := range targetEntitlements {
		existing[entitlement.LookupKey] = entitlement
	}

	total := len(entitlements)
	for i, entitlement := range entitlements {
		output.Progress(i+1, total, "Importing entitlement %s", entitlement.LookupKey)
		targetID := ""
		if found, ok := existing[entitlement.LookupKey]; ok {
			targetID = found.ID
			_, err := client.Post(fmt.Sprintf("/projects/%s/entitlements/%s", url.PathEscape(pid), url.PathEscape(targetID)), map[string]any{"display_name": entitlement.DisplayName})
			if err != nil {
				output.Warn("Failed to update entitlement %s: %v", entitlement.LookupKey, err)
			}
		} else {
			body := map[string]any{"lookup_key": entitlement.LookupKey, "display_name": entitlement.DisplayName}
			data, err := client.Post(fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(pid)), body)
			if err != nil {
				output.Warn("Failed to create entitlement %s: %v", entitlement.LookupKey, err)
				continue
			}
			var created api.Entitlement
			if err := json.Unmarshal(data, &created); err != nil {
				output.Warn("Failed to parse created entitlement %s: %v", entitlement.LookupKey, err)
				continue
			}
			targetID = created.ID
		}
		entitlementMap[entitlement.ID] = targetID
		result.Entitlements++
		attachEntitlementProducts(client, pid, targetID, entitlement, productMap)
	}
	return entitlementMap
}

func importOfferings(client *api.Client, pid string, offerings []OfferingConfig, targetOfferings []api.Offering, productMap map[string]string, result *importResult) map[string]string {
	offeringMap := make(map[string]string, len(offerings))
	existing := make(map[string]api.Offering, len(targetOfferings))
	for _, offering := range targetOfferings {
		existing[offering.LookupKey] = offering
	}

	total := len(offerings)
	for i, offering := range offerings {
		output.Progress(i+1, total, "Importing offering %s", offering.LookupKey)
		targetID := ""
		if found, ok := existing[offering.LookupKey]; ok {
			targetID = found.ID
			body := offeringUpdateBody(offering)
			if len(body) > 0 {
				_, err := client.Post(fmt.Sprintf("/projects/%s/offerings/%s", url.PathEscape(pid), url.PathEscape(targetID)), body)
				if err != nil {
					output.Warn("Failed to update offering %s: %v", offering.LookupKey, err)
				}
			}
		} else {
			body := offeringCreateBody(offering)
			data, err := client.Post(fmt.Sprintf("/projects/%s/offerings", url.PathEscape(pid)), body)
			if err != nil {
				output.Warn("Failed to create offering %s: %v", offering.LookupKey, err)
				continue
			}
			var created api.Offering
			if err := json.Unmarshal(data, &created); err != nil {
				output.Warn("Failed to parse created offering %s: %v", offering.LookupKey, err)
				continue
			}
			targetID = created.ID
		}
		offeringMap[offering.ID] = targetID
		result.Offerings++
		importPackages(client, pid, targetID, offering.Packages, productMap, result)
		if offering.IsCurrent {
			_, err := client.Post(fmt.Sprintf("/projects/%s/offerings/%s", url.PathEscape(pid), url.PathEscape(targetID)), map[string]any{"is_current": true})
			if err != nil {
				output.Warn("Failed to set offering %s as current: %v", offering.LookupKey, err)
			}
		}
	}
	return offeringMap
}

func importPackages(client *api.Client, pid, offeringID string, packages []PackageConfig, productMap map[string]string, result *importResult) {
	targetPackages, err := api.PaginateAll[api.Package](client, fmt.Sprintf("/projects/%s/offerings/%s/packages", url.PathEscape(pid), url.PathEscape(offeringID)), nil)
	if err != nil {
		output.Warn("Failed to fetch target packages for offering %s: %v", offeringID, err)
		return
	}
	existing := make(map[string]api.Package, len(targetPackages))
	for _, pkg := range targetPackages {
		existing[pkg.LookupKey] = pkg
	}

	total := len(packages)
	for i, pkg := range packages {
		output.Progress(i+1, total, "Importing package %s", pkg.LookupKey)
		targetID := ""
		if found, ok := existing[pkg.LookupKey]; ok {
			targetID = found.ID
			_, err := client.Post(fmt.Sprintf("/projects/%s/packages/%s", url.PathEscape(pid), url.PathEscape(targetID)), packageUpdateBody(pkg))
			if err != nil {
				output.Warn("Failed to update package %s: %v", pkg.LookupKey, err)
			}
		} else {
			data, err := client.Post(fmt.Sprintf("/projects/%s/offerings/%s/packages", url.PathEscape(pid), url.PathEscape(offeringID)), packageCreateBody(pkg))
			if err != nil {
				output.Warn("Failed to create package %s: %v", pkg.LookupKey, err)
				continue
			}
			var created api.Package
			if err := json.Unmarshal(data, &created); err != nil {
				output.Warn("Failed to parse created package %s: %v", pkg.LookupKey, err)
				continue
			}
			targetID = created.ID
		}
		result.Packages++
		attachPackageProducts(client, pid, targetID, pkg, productMap)
	}
}

func attachEntitlementProducts(client *api.Client, pid, targetEntitlementID string, entitlement EntitlementConfig, productMap map[string]string) {
	targetProductIDs := make([]string, 0, len(entitlement.ProductIDs))
	for _, sourceProductID := range entitlement.ProductIDs {
		targetProductID, ok := productMap[sourceProductID]
		if !ok {
			output.Warn("Skipping entitlement product attachment %s -> %s: product was not imported", entitlement.LookupKey, sourceProductID)
			continue
		}
		targetProductIDs = append(targetProductIDs, targetProductID)
	}
	if len(targetProductIDs) == 0 {
		return
	}
	_, err := client.Post(
		fmt.Sprintf("/projects/%s/entitlements/%s/actions/attach_products", url.PathEscape(pid), url.PathEscape(targetEntitlementID)),
		map[string]any{"product_ids": targetProductIDs},
	)
	if err != nil {
		output.Warn("Failed to attach products to entitlement %s: %v", entitlement.LookupKey, err)
	}
}

func attachPackageProducts(client *api.Client, pid, targetPackageID string, pkg PackageConfig, productMap map[string]string) {
	targetProducts := make([]map[string]any, 0, len(pkg.Products))
	for _, product := range pkg.Products {
		targetProductID, ok := productMap[product.ProductID]
		if !ok {
			output.Warn("Skipping package product attachment %s -> %s: product was not imported", pkg.LookupKey, product.ProductID)
			continue
		}
		eligibility := product.EligibilityCriteria
		if eligibility == "" {
			eligibility = "all"
		}
		targetProducts = append(targetProducts, map[string]any{
			"product_id":           targetProductID,
			"eligibility_criteria": eligibility,
		})
	}
	if len(targetProducts) == 0 {
		return
	}
	_, err := client.Post(
		fmt.Sprintf("/projects/%s/packages/%s/actions/attach_products", url.PathEscape(pid), url.PathEscape(targetPackageID)),
		map[string]any{"products": targetProducts},
	)
	if err != nil {
		output.Warn("Failed to attach products to package %s: %v", pkg.LookupKey, err)
	}
}

func archiveImportedEntities(client *api.Client, pid string, products []api.Product, entitlements []EntitlementConfig, offerings []OfferingConfig, productMap, entitlementMap, offeringMap map[string]string) {
	for _, offering := range offerings {
		if offering.State != "archived" {
			continue
		}
		targetID := offeringMap[offering.ID]
		if targetID == "" {
			continue
		}
		if _, err := client.Post(fmt.Sprintf("/projects/%s/offerings/%s/actions/archive", url.PathEscape(pid), url.PathEscape(targetID)), nil); err != nil {
			output.Warn("Failed to archive offering %s: %v", offering.LookupKey, err)
		}
	}
	for _, entitlement := range entitlements {
		if entitlement.State != "archived" {
			continue
		}
		targetID := entitlementMap[entitlement.ID]
		if targetID == "" {
			continue
		}
		if _, err := client.Post(fmt.Sprintf("/projects/%s/entitlements/%s/actions/archive", url.PathEscape(pid), url.PathEscape(targetID)), nil); err != nil {
			output.Warn("Failed to archive entitlement %s: %v", entitlement.LookupKey, err)
		}
	}
	for _, product := range products {
		if product.State != "archived" {
			continue
		}
		targetID := productMap[product.ID]
		if targetID == "" {
			continue
		}
		if _, err := client.Post(fmt.Sprintf("/projects/%s/products/%s/actions/archive", url.PathEscape(pid), url.PathEscape(targetID)), nil); err != nil {
			output.Warn("Failed to archive product %s: %v", product.StoreIdentifier, err)
		}
	}
}

func offeringCreateBody(offering OfferingConfig) map[string]any {
	body := map[string]any{
		"lookup_key":   offering.LookupKey,
		"display_name": offering.DisplayName,
	}
	if len(offering.Metadata) > 0 {
		body["metadata"] = offering.Metadata
	}
	return body
}

func offeringUpdateBody(offering OfferingConfig) map[string]any {
	body := map[string]any{"display_name": offering.DisplayName}
	if len(offering.Metadata) > 0 {
		body["metadata"] = offering.Metadata
	}
	return body
}

func packageCreateBody(pkg PackageConfig) map[string]any {
	body := map[string]any{"lookup_key": pkg.LookupKey, "display_name": pkg.DisplayName}
	if pkg.Position != nil {
		body["position"] = *pkg.Position
	}
	return body
}

func packageUpdateBody(pkg PackageConfig) map[string]any {
	body := map[string]any{"display_name": pkg.DisplayName}
	if pkg.Position != nil {
		body["position"] = *pkg.Position
	}
	return body
}

func appKey(name, typ string) string {
	return name + "\x00" + typ
}

func productKey(appID, storeIdentifier, typ string) string {
	return appID + "\x00" + storeIdentifier + "\x00" + typ
}

func countPackages(offerings []OfferingConfig) int {
	total := 0
	for _, offering := range offerings {
		total += len(offering.Packages)
	}
	return total
}
