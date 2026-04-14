package transfer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/output"
	"github.com/spf13/cobra"
)

// ProjectConfig is the top-level export format for a full project configuration.
type ProjectConfig struct {
	Version      string             `json:"version"`
	ExportedAt   string             `json:"exported_at"`
	Products     []api.Product      `json:"products"`
	Entitlements []api.Entitlement  `json:"entitlements"`
	Offerings    []api.Offering     `json:"offerings"`
}

// NewExportCmd creates the `rc export` command.
func NewExportCmd(projectID, outputFormat *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export full project configuration (products, entitlements, offerings)",
		Long: `Export the complete project configuration as a single JSON file.
This includes all products, entitlements, and offerings with their packages.

Example:
  rc export --file config.json`,
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

			// Fetch products
			prodData, err := client.Get(fmt.Sprintf("/projects/%s/products", url.PathEscape(pid)), nil)
			if err != nil {
				return fmt.Errorf("failed to fetch products: %w", err)
			}
			var prodResp api.ListResponse[api.Product]
			if err := json.Unmarshal(prodData, &prodResp); err != nil {
				return fmt.Errorf("failed to parse products: %w", err)
			}

			// Fetch entitlements
			entData, err := client.Get(fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(pid)), nil)
			if err != nil {
				return fmt.Errorf("failed to fetch entitlements: %w", err)
			}
			var entResp api.ListResponse[api.Entitlement]
			if err := json.Unmarshal(entData, &entResp); err != nil {
				return fmt.Errorf("failed to parse entitlements: %w", err)
			}

			// Fetch offerings with packages expanded
			query := url.Values{}
			query.Set("expand", "items.package,items.package.product")
			offData, err := client.Get(fmt.Sprintf("/projects/%s/offerings", url.PathEscape(pid)), query)
			if err != nil {
				return fmt.Errorf("failed to fetch offerings: %w", err)
			}
			var offResp api.ListResponse[api.Offering]
			if err := json.Unmarshal(offData, &offResp); err != nil {
				return fmt.Errorf("failed to parse offerings: %w", err)
			}

			config := ProjectConfig{
				Version:      "1",
				ExportedAt:   time.Now().UTC().Format(time.RFC3339),
				Products:     prodResp.Items,
				Entitlements: entResp.Items,
				Offerings:    offResp.Items,
			}

			data, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			if err := os.WriteFile(file, data, 0644); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			output.Success("Exported %d products, %d entitlements, %d offerings to %s",
				len(config.Products), len(config.Entitlements), len(config.Offerings), file)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "output file path (required)")
	return cmd
}

// NewImportCmd creates the `rc import` command.
func NewImportCmd(projectID, outputFormat *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import project configuration from a JSON file",
		Long: `Import a project configuration exported with 'rc export'.
Creates products, entitlements, and offerings in the target project.

Example:
  rc import --file config.json --project proj_target123`,
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

			// Create products
			prodCreated := 0
			for _, p := range config.Products {
				body := map[string]any{
					"store_identifier": p.StoreIdentifier,
					"app_id":           p.AppID,
					"type":             p.Type,
				}
				if p.DisplayName != nil {
					body["display_name"] = *p.DisplayName
				}
				_, err := client.Post(fmt.Sprintf("/projects/%s/products", url.PathEscape(pid)), body)
				if err != nil {
					output.Warn("Failed to create product %s: %v", p.StoreIdentifier, err)
					continue
				}
				prodCreated++
			}

			// Create entitlements
			entCreated := 0
			for _, e := range config.Entitlements {
				body := map[string]any{
					"lookup_key":   e.LookupKey,
					"display_name": e.DisplayName,
				}
				_, err := client.Post(fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(pid)), body)
				if err != nil {
					output.Warn("Failed to create entitlement %s: %v", e.LookupKey, err)
					continue
				}
				entCreated++
			}

			// Create offerings
			offCreated := 0
			for _, o := range config.Offerings {
				body := map[string]any{
					"lookup_key":   o.LookupKey,
					"display_name": o.DisplayName,
				}
				_, err := client.Post(fmt.Sprintf("/projects/%s/offerings", url.PathEscape(pid)), body)
				if err != nil {
					output.Warn("Failed to create offering %s: %v", o.LookupKey, err)
					continue
				}
				offCreated++
			}

			output.Success("Imported %d/%d products, %d/%d entitlements, %d/%d offerings",
				prodCreated, len(config.Products),
				entCreated, len(config.Entitlements),
				offCreated, len(config.Offerings))
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "input file path (required)")
	return cmd
}
