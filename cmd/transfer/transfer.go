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
