package entitlements

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

func NewEntitlementsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "entitlements",
		Aliases: []string{"entitlement", "ent"},
		Short:   "Manage entitlements",
		Long: `Manage entitlements in a RevenueCat project.

Entitlements represent levels of access that a customer can "unlock".
Products are attached to entitlements to grant access when purchased.

Examples:
  rc entitlements list
  rc entitlements get entla1b2c3d4e5
  rc entitlements create --lookup-key premium --display-name "Premium"
  rc entitlements attach --entitlement-id entla1b2c3d4e5 --product-id prod1a2b3c4d5e
  rc entitlements archive entla1b2c3d4e5`,
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newGetCmd(projectID, outputFormat))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	root.AddCommand(newUpdateCmd(projectID, outputFormat))
	root.AddCommand(newDeleteCmd(projectID))
	root.AddCommand(newArchiveCmd(projectID))
	root.AddCommand(newUnarchiveCmd(projectID))
	root.AddCommand(newListProductsCmd(projectID, outputFormat))
	root.AddCommand(newAttachCmd(projectID))
	root.AddCommand(newDetachCmd(projectID))
	root.AddCommand(newExportCmd(projectID))
	root.AddCommand(newImportCmd(projectID))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		fetchAll bool
		limit    int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List entitlements in a project",
		Example: `  # List all entitlements
  rc entitlements list

  # List entitlements for a specific project as JSON
  rc entitlements list --project proj1a2b3c4d5 --output json

  # Use a production profile
  rc entitlements list --profile production

  # Extract lookup keys for documentation
  rc entitlements list --output json | jq -r '.items[].lookup_key'

  # Find an entitlement, then list attached products
  rc entitlements list --output json | jq -r '.items[0].id'
  rc entitlements products entla1b2c3d4e5

  # Fetch every page
  rc entitlements list --all --limit 100`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			path := fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(pid))
			query := url.Values{}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}

			if fetchAll {
				items, err := api.PaginateAll[api.Entitlement](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Lookup Key", "Display Name", "State", "Created"})
					for _, e := range items {
						t.AppendRow(table.Row{e.ID, e.LookupKey, e.DisplayName, e.State, output.FormatTimestamp(e.CreatedAt)})
					}
					t.AppendFooter(table.Row{"", "", "", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}

			data, err := client.Get(path, query)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.Entitlement]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Lookup Key", "Display Name", "State", "Created"})
				for _, e := range resp.Items {
					t.AppendRow(table.Row{e.ID, e.LookupKey, e.DisplayName, e.State, output.FormatTimestamp(e.CreatedAt)})
				}
			})
			if resp.NextPage != nil {
				output.Warn("More results available (use --all for more)")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&fetchAll, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&limit, "limit", 0, "max items per page")
	return cmd
}

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <entitlement-id>",
		Short: "Get an entitlement by ID",
		Example: `  # Get entitlement details
  rc entitlements get entla1b2c3d4e5

  # Get as JSON
  rc entitlements get entla1b2c3d4e5 -o json`,
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

			data, err := client.Get(fmt.Sprintf("/projects/%s/entitlements/%s", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}

			var ent api.Entitlement
			if err := json.Unmarshal(data, &ent); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, ent, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", ent.ID},
					{"Lookup Key", ent.LookupKey},
					{"Display Name", ent.DisplayName},
					{"State", ent.State},
					{"Created", output.FormatTimestamp(ent.CreatedAt)},
				})
			})
			return nil
		},
	}
}

func newCreateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		lookupKey   string
		displayName string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new entitlement",
		Long: `Create a new entitlement. Required flags are prompted interactively when
running in a terminal and not provided on the command line.`,
		Example: `  # Create an entitlement
  rc entitlements create --lookup-key premium --display-name "Premium Access"

  # Create and print JSON
  rc entitlements create --lookup-key pro --display-name "Pro Access" --output json

  # Use a staging profile
  rc entitlements create --lookup-key beta --display-name "Beta Access" --profile staging

  # Capture the entitlement ID for a workflow
  rc entitlements create --lookup-key premium --display-name "Premium Access" --output json | jq -r '.id'

  # Create, then attach a product
  rc entitlements create --lookup-key premium --display-name "Premium Access"
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1a2b3c

  # Interactive mode (prompts for missing fields)
  rc entitlements create`,
		RunE: func(c *cobra.Command, args []string) error {
			// Interactive prompts for missing required fields
			if err := cmdutil.PromptIfEmpty(&lookupKey, "Lookup key", "premium"); err != nil {
				return err
			}
			if err := cmdutil.PromptIfEmpty(&displayName, "Display name", "Premium Access"); err != nil {
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

			body := map[string]any{"lookup_key": lookupKey, "display_name": displayName}
			data, err := client.Post(fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(pid)), body)
			if err != nil {
				return err
			}

			var ent api.Entitlement
			if err := json.Unmarshal(data, &ent); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, ent, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", ent.ID},
					{"Lookup Key", ent.LookupKey},
					{"Display Name", ent.DisplayName},
					{"State", ent.State},
					{"Created", output.FormatTimestamp(ent.CreatedAt)},
				})
			})
			output.Success("Entitlement created successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&lookupKey, "lookup-key", "", "lookup key identifier")
	cmd.Flags().StringVar(&displayName, "display-name", "", "display name")
	return cmd
}

func newUpdateCmd(projectID, outputFormat *string) *cobra.Command {
	var displayName string

	cmd := &cobra.Command{
		Use:   "update <entitlement-id>",
		Short: "Update an entitlement",
		Example: `  # Update display name
  rc entitlements update entla1b2c3d4e5 --display-name "Premium Plus"`,
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

			body := map[string]any{"display_name": displayName}
			data, err := client.Post(fmt.Sprintf("/projects/%s/entitlements/%s", url.PathEscape(pid), url.PathEscape(args[0])), body)
			if err != nil {
				return err
			}

			var ent api.Entitlement
			if err := json.Unmarshal(data, &ent); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, ent, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", ent.ID},
					{"Display Name", ent.DisplayName},
					{"State", ent.State},
				})
			})
			output.Success("Entitlement updated")
			return nil
		},
	}

	cmd.Flags().StringVar(&displayName, "display-name", "", "new display name (required)")
	cmdutil.MustMarkFlagRequired(cmd, "display-name")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <entitlement-id>",
		Short: "Delete an entitlement",
		Example: `  # Delete an entitlement
  rc entitlements delete entla1b2c3d4e5`,
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
			_, err = client.Delete(fmt.Sprintf("/projects/%s/entitlements/%s", url.PathEscape(pid), url.PathEscape(args[0])))
			if err != nil {
				return err
			}
			output.Success("Entitlement %s deleted", args[0])
			return nil
		},
	}
}

func newArchiveCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <entitlement-id>",
		Short: "Archive an entitlement",
		Example: `  # Archive an entitlement
  rc entitlements archive entla1b2c3d4e5`,
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
			_, err = client.Post(fmt.Sprintf("/projects/%s/entitlements/%s/actions/archive", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Entitlement %s archived", args[0])
			return nil
		},
	}
}

func newUnarchiveCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "unarchive <entitlement-id>",
		Short: "Unarchive an entitlement",
		Example: `  # Unarchive an entitlement
  rc entitlements unarchive entla1b2c3d4e5`,
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
			_, err = client.Post(fmt.Sprintf("/projects/%s/entitlements/%s/actions/unarchive", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Entitlement %s unarchived", args[0])
			return nil
		},
	}
}

func newListProductsCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		fetchAll bool
		limit    int
	)

	cmd := &cobra.Command{
		Use:   "products <entitlement-id>",
		Short: "List products attached to an entitlement",
		Example: `  # List products for an entitlement
  rc entitlements products entla1b2c3d4e5

  # Fetch all pages
  rc entitlements products entla1b2c3d4e5 --all`,
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

			path := fmt.Sprintf("/projects/%s/entitlements/%s/products", url.PathEscape(pid), url.PathEscape(args[0]))
			query := url.Values{}
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
					t.AppendHeader(table.Row{"ID", "Store ID", "Type", "State"})
					for _, p := range items {
						t.AppendRow(table.Row{p.ID, p.StoreIdentifier, p.Type, p.State})
					}
					t.AppendFooter(table.Row{"", "", "", fmt.Sprintf("%d total", len(items))})
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
				t.AppendHeader(table.Row{"ID", "Store ID", "Type", "State"})
				for _, p := range resp.Items {
					t.AppendRow(table.Row{p.ID, p.StoreIdentifier, p.Type, p.State})
				}
			})
			if resp.NextPage != nil {
				output.Warn("More results available (use --all for more)")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&fetchAll, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&limit, "limit", 0, "max items per page")
	return cmd
}

func newAttachCmd(projectID *string) *cobra.Command {
	var (
		entitlementID string
		productIDs    []string
	)

	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach products to an entitlement",
		Long:  `Attach one or more products to an entitlement.`,
		Example: `  # Attach a single product
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1a2b3c

  # Attach multiple products
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1,prod2

  # Use a production profile
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1a2b3c --profile production

  # Attach the first product returned from a query
  rc products list --output json | jq -r '.items[0].id'
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1a2b3c

  # Verify attached products
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1a2b3c
  rc entitlements products entla1b2c3 --output json`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			_, err = client.Post(
				fmt.Sprintf("/projects/%s/entitlements/%s/actions/attach_products", url.PathEscape(pid), url.PathEscape(entitlementID)),
				map[string]any{"product_ids": productIDs},
			)
			if err != nil {
				return err
			}
			output.Success("Attached %d product(s) to entitlement %s", len(productIDs), entitlementID)
			return nil
		},
	}

	cmd.Flags().StringVar(&entitlementID, "entitlement-id", "", "entitlement ID (required)")
	cmd.Flags().StringSliceVar(&productIDs, "product-id", nil, "product ID(s) to attach (required, comma-separated)")
	cmdutil.MustMarkFlagRequired(cmd, "entitlement-id")
	cmdutil.MustMarkFlagRequired(cmd, "product-id")
	return cmd
}

func newDetachCmd(projectID *string) *cobra.Command {
	var (
		entitlementID string
		productIDs    []string
	)

	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach products from an entitlement",
		Example: `  # Detach a product
  rc entitlements detach --entitlement-id entla1b2c3 --product-id prod1a2b3c

  # Detach multiple products
  rc entitlements detach --entitlement-id entla1b2c3 --product-id prod1,prod2`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			_, err = client.Post(
				fmt.Sprintf("/projects/%s/entitlements/%s/actions/detach_products", url.PathEscape(pid), url.PathEscape(entitlementID)),
				map[string]any{"product_ids": productIDs},
			)
			if err != nil {
				return err
			}
			output.Success("Detached %d product(s) from entitlement %s", len(productIDs), entitlementID)
			return nil
		},
	}

	cmd.Flags().StringVar(&entitlementID, "entitlement-id", "", "entitlement ID (required)")
	cmd.Flags().StringSliceVar(&productIDs, "product-id", nil, "product ID(s) to detach (required, comma-separated)")
	cmdutil.MustMarkFlagRequired(cmd, "entitlement-id")
	cmdutil.MustMarkFlagRequired(cmd, "product-id")
	return cmd
}

// EntitlementRow is a flat CSV row for entitlement export/import.
type EntitlementRow struct {
	LookupKey   string `csv:"lookup_key" json:"lookup_key"`
	DisplayName string `csv:"display_name" json:"display_name"`
}

func newExportCmd(projectID *string) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export entitlements to a file (CSV or JSON)",
		Long:  `Export entitlements to a CSV or JSON file. Format is detected from the file extension.`,
		Example: `  # Export to CSV
  rc entitlements export --file entitlements.csv

  # Export to JSON
  rc entitlements export --file entitlements.json`,
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

			data, err := client.Get(fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(pid)), nil)
			if err != nil {
				return err
			}
			var resp api.ListResponse[api.Entitlement]
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
				rows := make([]EntitlementRow, len(resp.Items))
				for i, e := range resp.Items {
					rows[i] = EntitlementRow{
						LookupKey:   e.LookupKey,
						DisplayName: e.DisplayName,
					}
				}
				f, err := os.Create(file)
				if err != nil {
					return err
				}
				if err := csvio.ExportCSV(f, rows); err != nil {
					if closeErr := f.Close(); closeErr != nil {
						return closeErr
					}
					return err
				}
				if err := f.Close(); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported file extension %q (use .csv or .json)", ext)
			}

			output.Success("Exported %d entitlements to %s", len(resp.Items), file)
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
		Short: "Import entitlements from a file (CSV or JSON)",
		Long: `Import entitlements from a CSV or JSON file. Format is detected from the file extension.
Each row/entry creates a new entitlement in the project.`,
		Example: `  # Import from CSV
  rc entitlements import --file entitlements.csv

  # Import from JSON
  rc entitlements import --file entitlements.json`,
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

			var rows []EntitlementRow
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
				rows, err = csvio.ImportCSV[EntitlementRow](f)
				if err != nil {
					if closeErr := f.Close(); closeErr != nil {
						return closeErr
					}
					return fmt.Errorf("failed to parse CSV: %w", err)
				}
				if err := f.Close(); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported file extension %q (use .csv or .json)", ext)
			}

			created := 0
			total := len(rows)
			for i, row := range rows {
				output.Progress(i+1, total, "Creating entitlement %s", row.LookupKey)
				body := map[string]any{
					"lookup_key":   row.LookupKey,
					"display_name": row.DisplayName,
				}
				_, err := client.Post(fmt.Sprintf("/projects/%s/entitlements", url.PathEscape(pid)), body)
				if err != nil {
					output.Warn("Failed to create entitlement %s: %v", row.LookupKey, err)
					continue
				}
				created++
			}

			output.Success("Imported %d/%d entitlements", created, len(rows))
			return nil
		},
	}

	cmd.Flags().StringVar(&file, "file", "", "input file path (.csv or .json)")
	return cmd
}
