package products

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/csvio"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewProductsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "products",
		Aliases: []string{"product", "prod"},
		Short:   "Manage products",
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newGetCmd(projectID, outputFormat))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	root.AddCommand(newUpdateCmd(projectID, outputFormat))
	root.AddCommand(newDeleteCmd(projectID))
	root.AddCommand(newArchiveCmd(projectID))
	root.AddCommand(newUnarchiveCmd(projectID))
	root.AddCommand(newPushToStoreCmd(projectID))
	root.AddCommand(newExportCmd(projectID))
	root.AddCommand(newImportCmd(projectID))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		appID    string
		fetchAll bool
		limit    int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List products in a project",
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			path := fmt.Sprintf("/projects/%s/products", url.PathEscape(pid))
			query := url.Values{}
			if appID != "" {
				query.Set("app_id", appID)
			}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}

			if fetchAll {
				items, err := api.PaginateAll[api.Product](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Store ID", "Type", "State", "Display Name", "Created"})
					for _, p := range items {
						t.AppendRow(table.Row{p.ID, p.StoreIdentifier, p.Type, p.State, output.Deref(p.DisplayName, "-"), output.FormatTimestamp(p.CreatedAt)})
					}
					t.AppendFooter(table.Row{"", "", "", "", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}

			data, err := client.Get(path, query)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.Product]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Store ID", "Type", "State", "Display Name", "Created"})
				for _, p := range resp.Items {
					t.AppendRow(table.Row{p.ID, p.StoreIdentifier, p.Type, p.State, output.Deref(p.DisplayName, "-"), output.FormatTimestamp(p.CreatedAt)})
				}
			})
			if resp.NextPage != nil {
				output.Warn("More results available (use --all for more)")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "filter by app ID")
	cmd.Flags().BoolVar(&fetchAll, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&limit, "limit", 0, "max items per page")
	return cmd
}

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <product-id>",
		Short: "Get a product by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			data, err := client.Get(fmt.Sprintf("/projects/%s/products/%s", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}

			var product api.Product
			if err := json.Unmarshal(data, &product); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, product, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", product.ID},
					{"Store ID", product.StoreIdentifier},
					{"Type", product.Type},
					{"State", product.State},
					{"Display Name", output.Deref(product.DisplayName, "-")},
					{"App ID", product.AppID},
					{"Created", output.FormatTimestamp(product.CreatedAt)},
				})
			})
			return nil
		},
	}
}

func newCreateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		storeIdentifier string
		appID           string
		productType     string
		displayName     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new product",
		Long: `Create a new product. Required flags are prompted interactively when
running in a terminal and not provided on the command line.`,
		RunE: func(c *cobra.Command, args []string) error {
			// Interactive prompts for missing required fields
			if err := cmdutil.PromptIfEmpty(&storeIdentifier, "Store identifier", "com.app.product_id"); err != nil {
				return err
			}
			if err := cmdutil.PromptIfEmpty(&appID, "App ID", "app1a2b3c4d5e"); err != nil {
				return err
			}
			if err := cmdutil.PromptSelect(&productType, "Product type", []string{"subscription", "one_time", "consumable", "non_consumable"}); err != nil {
				return err
			}

			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			body := map[string]any{
				"store_identifier": storeIdentifier,
				"app_id":           appID,
				"type":             productType,
			}
			if displayName != "" {
				body["display_name"] = displayName
			}

			data, err := client.Post(fmt.Sprintf("/projects/%s/products", url.PathEscape(pid)), body)
			if err != nil {
				return err
			}

			var product api.Product
			if err := json.Unmarshal(data, &product); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, product, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", product.ID},
					{"Store ID", product.StoreIdentifier},
					{"Type", product.Type},
					{"State", product.State},
					{"Display Name", output.Deref(product.DisplayName, "-")},
					{"App ID", product.AppID},
					{"Created", output.FormatTimestamp(product.CreatedAt)},
				})
			})
			output.Success("Product created successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&storeIdentifier, "store-id", "", "store product identifier")
	cmd.Flags().StringVar(&appID, "app-id", "", "app ID")
	cmd.Flags().StringVar(&productType, "type", "", "product type: subscription, one_time, consumable, non_consumable")
	cmd.Flags().StringVar(&displayName, "display-name", "", "display name")
	return cmd
}

func newUpdateCmd(projectID, outputFormat *string) *cobra.Command {
	var displayName string

	cmd := &cobra.Command{
		Use:   "update <product-id>",
		Short: "Update a product",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			body := map[string]any{"display_name": displayName}
			data, err := client.Post(fmt.Sprintf("/projects/%s/products/%s", url.PathEscape(pid), url.PathEscape(args[0])), body)
			if err != nil {
				return err
			}

			var product api.Product
			if err := json.Unmarshal(data, &product); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, product, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", product.ID},
					{"Display Name", output.Deref(product.DisplayName, "-")},
					{"State", product.State},
				})
			})
			output.Success("Product updated")
			return nil
		},
	}

	cmd.Flags().StringVar(&displayName, "display-name", "", "new display name (required)")
	cmd.MarkFlagRequired("display-name")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <product-id>",
		Short: "Delete a product",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			_, err = client.Delete(fmt.Sprintf("/projects/%s/products/%s", url.PathEscape(pid), url.PathEscape(args[0])))
			if err != nil {
				return err
			}
			output.Success("Product %s deleted", args[0])
			return nil
		},
	}
}

func newArchiveCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <product-id>",
		Short: "Archive a product",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			_, err = client.Post(fmt.Sprintf("/projects/%s/products/%s/actions/archive", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Product %s archived", args[0])
			return nil
		},
	}
}

func newUnarchiveCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "unarchive <product-id>",
		Short: "Unarchive a product",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			_, err = client.Post(fmt.Sprintf("/projects/%s/products/%s/actions/unarchive", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Product %s unarchived", args[0])
			return nil
		},
	}
}

func newPushToStoreCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "push-to-store <product-id>",
		Short: "Push a product to the store (create in the connected store)",
		Long: `Push a product configuration to the connected app store.

This creates the product in the store (e.g., App Store Connect, Google Play)
using the product configuration defined in RevenueCat.`,
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
			_, err = client.Post(fmt.Sprintf("/projects/%s/products/%s/create_in_store", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Product %s pushed to store", args[0])
			return nil
		},
	}
}

// ProductRow is a flat CSV row for product export/import.
type ProductRow struct {
	StoreIdentifier string `csv:"store_identifier" json:"store_identifier"`
	AppID           string `csv:"app_id" json:"app_id"`
	Type            string `csv:"type" json:"type"`
	DisplayName     string `csv:"display_name" json:"display_name"`
}

func newExportCmd(projectID *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export products to a file (CSV or JSON)",
		Long: `Export products to a CSV or JSON file. Format is detected from the file extension.

Examples:
  rc products export --file products.csv
  rc products export --file products.json`,
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

			data, err := client.Get(fmt.Sprintf("/projects/%s/products", url.PathEscape(pid)), nil)
			if err != nil {
				return err
			}
			var resp api.ListResponse[api.Product]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			ext := strings.ToLower(filepath.Ext(file))
			switch ext {
			case ".json":
				out, err := json.MarshalIndent(resp.Items, "", "  ")
				if err != nil {
					return err
				}
				if err := os.WriteFile(file, out, 0644); err != nil {
					return err
				}
			case ".csv":
				rows := make([]ProductRow, len(resp.Items))
				for i, p := range resp.Items {
					rows[i] = ProductRow{
						StoreIdentifier: p.StoreIdentifier,
						AppID:           p.AppID,
						Type:            p.Type,
						DisplayName:     output.Deref(p.DisplayName, ""),
					}
				}
				f, err := os.Create(file)
				if err != nil {
					return err
				}
				defer f.Close()
				if err := csvio.ExportCSV(f, rows); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported file extension %q (use .csv or .json)", ext)
			}

			output.Success("Exported %d products to %s", len(resp.Items), file)
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "output file path (.csv or .json)")
	return cmd
}

func newImportCmd(projectID *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import products from a file (CSV or JSON)",
		Long: `Import products from a CSV or JSON file. Format is detected from the file extension.
Each row/entry creates a new product in the project.

Examples:
  rc products import --file products.csv
  rc products import --file products.json`,
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

			var rows []ProductRow
			ext := strings.ToLower(filepath.Ext(file))
			switch ext {
			case ".json":
				data, err := os.ReadFile(file)
				if err != nil {
					return err
				}
				if err := json.Unmarshal(data, &rows); err != nil {
					return fmt.Errorf("failed to parse JSON: %w", err)
				}
			case ".csv":
				f, err := os.Open(file)
				if err != nil {
					return err
				}
				defer f.Close()
				rows, err = csvio.ImportCSV[ProductRow](f)
				if err != nil {
					return fmt.Errorf("failed to parse CSV: %w", err)
				}
			default:
				return fmt.Errorf("unsupported file extension %q (use .csv or .json)", ext)
			}

			created := 0
			for _, row := range rows {
				body := map[string]any{
					"store_identifier": row.StoreIdentifier,
					"app_id":           row.AppID,
					"type":             row.Type,
				}
				if row.DisplayName != "" {
					body["display_name"] = row.DisplayName
				}
				_, err := client.Post(fmt.Sprintf("/projects/%s/products", url.PathEscape(pid)), body)
				if err != nil {
					output.Warn("Failed to create product %s: %v", row.StoreIdentifier, err)
					continue
				}
				created++
			}

			output.Success("Imported %d/%d products", created, len(rows))
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "input file path (.csv or .json)")
	return cmd
}
