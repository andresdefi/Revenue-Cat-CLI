package packages

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

func NewPackagesCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "packages",
		Aliases: []string{"package", "pkg"},
		Short:   "Manage packages within offerings",
		Long: `Manage packages within RevenueCat offerings.

Packages are the unit that ties products to offerings. Each package in an
offering can have one or more products attached (for different platforms).

Examples:
  rc packages list --offering-id ofrnge1a2b3c
  rc packages get pkge1a2b3c4d5
  rc packages create --offering-id ofrnge1a2b3c --lookup-key monthly --display-name "Monthly"
  rc packages attach --package-id pkge1a2b3c --product-id prod1a2b3c
  rc packages delete pkge1a2b3c4d5`,
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newGetCmd(projectID, outputFormat))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	root.AddCommand(newUpdateCmd(projectID, outputFormat))
	root.AddCommand(newDeleteCmd(projectID))
	root.AddCommand(newListProductsCmd(projectID, outputFormat))
	root.AddCommand(newAttachCmd(projectID))
	root.AddCommand(newDetachCmd(projectID))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		offeringID string
		fetchAll   bool
		limit      int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List packages in an offering",
		Example: `  # List packages in an offering
  rc packages list --offering-id ofrnge1a2b3c

  # List with JSON output
  rc packages list --offering-id ofrnge1a2b3c -o json

  # Fetch all pages
  rc packages list --offering-id ofrnge1a2b3c --all`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			path := fmt.Sprintf("/projects/%s/offerings/%s/packages", url.PathEscape(pid), url.PathEscape(offeringID))
			query := url.Values{}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}

			if fetchAll {
				items, err := api.PaginateAll[api.Package](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Lookup Key", "Display Name", "Position", "Created"})
					for _, p := range items {
						pos := "-"
						if p.Position != nil {
							pos = fmt.Sprintf("%d", *p.Position)
						}
						t.AppendRow(table.Row{p.ID, p.LookupKey, p.DisplayName, pos, output.FormatTimestamp(p.CreatedAt)})
					}
					t.AppendFooter(table.Row{"", "", "", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}

			data, err := client.Get(path, query)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.Package]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Lookup Key", "Display Name", "Position", "Created"})
				for _, p := range resp.Items {
					pos := "-"
					if p.Position != nil {
						pos = fmt.Sprintf("%d", *p.Position)
					}
					t.AppendRow(table.Row{p.ID, p.LookupKey, p.DisplayName, pos, output.FormatTimestamp(p.CreatedAt)})
				}
			})
			if resp.NextPage != nil {
				output.Warn("More results available (use --all for more)")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&offeringID, "offering-id", "", "offering ID (required)")
	cmdutil.MustMarkFlagRequired(cmd, "offering-id")
	cmd.Flags().BoolVar(&fetchAll, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&limit, "limit", 0, "max items per page")
	return cmd
}

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <package-id>",
		Short: "Get a package by ID",
		Example: `  # Get package details
  rc packages get pkge1a2b3c4d5

  # Get as JSON
  rc packages get pkge1a2b3c4d5 -o json`,
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
			data, err := client.Get(fmt.Sprintf("/projects/%s/packages/%s", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			var pkg api.Package
			if err := json.Unmarshal(data, &pkg); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, pkg, func(t table.Writer) {
				pos := "-"
				if pkg.Position != nil {
					pos = fmt.Sprintf("%d", *pkg.Position)
				}
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", pkg.ID},
					{"Lookup Key", pkg.LookupKey},
					{"Display Name", pkg.DisplayName},
					{"Position", pos},
					{"Created", output.FormatTimestamp(pkg.CreatedAt)},
				})
			})
			return nil
		},
	}
}

func newCreateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		offeringID  string
		lookupKey   string
		displayName string
		position    int
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new package in an offering",
		Example: `  # Create a package
  rc packages create --offering-id ofrnge1a2b3c --lookup-key monthly --display-name "Monthly"

  # Create with position
  rc packages create --offering-id ofrnge1a2b3c --lookup-key annual --display-name "Annual" --position 1`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			body := map[string]any{"lookup_key": lookupKey, "display_name": displayName}
			if c.Flags().Changed("position") {
				body["position"] = position
			}
			data, err := client.Post(fmt.Sprintf("/projects/%s/offerings/%s/packages", url.PathEscape(pid), url.PathEscape(offeringID)), body)
			if err != nil {
				return err
			}
			var pkg api.Package
			if err := json.Unmarshal(data, &pkg); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, pkg, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", pkg.ID},
					{"Lookup Key", pkg.LookupKey},
					{"Display Name", pkg.DisplayName},
					{"Created", output.FormatTimestamp(pkg.CreatedAt)},
				})
			})
			output.Success("Package created successfully")
			return nil
		},
	}
	cmd.Flags().StringVar(&offeringID, "offering-id", "", "offering ID (required)")
	cmd.Flags().StringVar(&lookupKey, "lookup-key", "", "lookup key (required)")
	cmd.Flags().StringVar(&displayName, "display-name", "", "display name (required)")
	cmd.Flags().IntVar(&position, "position", 0, "display position (min 1)")
	cmdutil.MustMarkFlagRequired(cmd, "offering-id")
	cmdutil.MustMarkFlagRequired(cmd, "lookup-key")
	cmdutil.MustMarkFlagRequired(cmd, "display-name")
	return cmd
}

func newUpdateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		displayName string
		position    int
	)
	cmd := &cobra.Command{
		Use:   "update <package-id>",
		Short: "Update a package",
		Example: `  # Update display name
  rc packages update pkge1a2b3c4d5 --display-name "Annual Plan"

  # Update position
  rc packages update pkge1a2b3c4d5 --position 2`,
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
			body := map[string]any{}
			if c.Flags().Changed("display-name") {
				body["display_name"] = displayName
			}
			if c.Flags().Changed("position") {
				body["position"] = position
			}
			data, err := client.Post(fmt.Sprintf("/projects/%s/packages/%s", url.PathEscape(pid), url.PathEscape(args[0])), body)
			if err != nil {
				return err
			}
			var pkg api.Package
			if err := json.Unmarshal(data, &pkg); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, pkg, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{{"ID", pkg.ID}, {"Display Name", pkg.DisplayName}})
			})
			output.Success("Package updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&displayName, "display-name", "", "new display name")
	cmd.Flags().IntVar(&position, "position", 0, "new position")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use: "delete <package-id>", Short: "Delete a package",
		Example: `  # Delete a package
  rc packages delete pkge1a2b3c4d5`,
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
			_, err = client.Delete(fmt.Sprintf("/projects/%s/packages/%s", url.PathEscape(pid), url.PathEscape(args[0])))
			if err != nil {
				return err
			}
			output.Success("Package %s deleted", args[0])
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
		Use:   "products <package-id>",
		Short: "List products attached to a package",
		Example: `  # List products for a package
  rc packages products pkge1a2b3c4d5

  # Fetch all pages
  rc packages products pkge1a2b3c4d5 --all`,
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
			path := fmt.Sprintf("/projects/%s/packages/%s/products", url.PathEscape(pid), url.PathEscape(args[0]))
			query := url.Values{}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}
			if fetchAll {
				items, err := api.PaginateAll[api.PackageProduct](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"Product ID", "Eligibility"})
					for _, p := range items {
						t.AppendRow(table.Row{p.ProductID, p.EligibilityCriteria})
					}
					t.AppendFooter(table.Row{"", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}
			data, err := client.Get(path, query)
			if err != nil {
				return err
			}
			var resp api.ListResponse[api.PackageProduct]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"Product ID", "Eligibility"})
				for _, p := range resp.Items {
					t.AppendRow(table.Row{p.ProductID, p.EligibilityCriteria})
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
		packageID   string
		productID   string
		eligibility string
	)
	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach a product to a package",
		Long: `Attach a product to a package with eligibility criteria.

Eligibility options: all (default), google_sdk_lt_6, google_sdk_ge_6`,
		Example: `  # Attach a product to a package
  rc packages attach --package-id pkge1a2b3c --product-id prod1a2b3c

  # Attach with eligibility criteria
  rc packages attach --package-id pkge1a2b3c --product-id prod1a2b3c --eligibility google_sdk_ge_6`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			products := []map[string]any{{"product_id": productID, "eligibility_criteria": eligibility}}
			_, err = client.Post(
				fmt.Sprintf("/projects/%s/packages/%s/actions/attach_products", url.PathEscape(pid), url.PathEscape(packageID)),
				map[string]any{"products": products},
			)
			if err != nil {
				return err
			}
			output.Success("Product %s attached to package %s", productID, packageID)
			return nil
		},
	}
	cmd.Flags().StringVar(&packageID, "package-id", "", "package ID (required)")
	cmd.Flags().StringVar(&productID, "product-id", "", "product ID (required)")
	cmd.Flags().StringVar(&eligibility, "eligibility", "all", "eligibility criteria: all, google_sdk_lt_6, google_sdk_ge_6")
	cmdutil.MustMarkFlagRequired(cmd, "package-id")
	cmdutil.MustMarkFlagRequired(cmd, "product-id")
	return cmd
}

func newDetachCmd(projectID *string) *cobra.Command {
	var (
		packageID  string
		productIDs []string
	)
	cmd := &cobra.Command{
		Use: "detach", Short: "Detach products from a package",
		Example: `  # Detach a product
  rc packages detach --package-id pkge1a2b3c --product-id prod1a2b3c

  # Detach multiple products
  rc packages detach --package-id pkge1a2b3c --product-id prod1,prod2`,
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
				fmt.Sprintf("/projects/%s/packages/%s/actions/detach_products", url.PathEscape(pid), url.PathEscape(packageID)),
				map[string]any{"product_ids": productIDs},
			)
			if err != nil {
				return err
			}
			output.Success("Detached %d product(s) from package %s", len(productIDs), packageID)
			return nil
		},
	}
	cmd.Flags().StringVar(&packageID, "package-id", "", "package ID (required)")
	cmd.Flags().StringSliceVar(&productIDs, "product-id", nil, "product ID(s) to detach (required, comma-separated)")
	cmdutil.MustMarkFlagRequired(cmd, "package-id")
	cmdutil.MustMarkFlagRequired(cmd, "product-id")
	return cmd
}
