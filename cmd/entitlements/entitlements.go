package entitlements

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
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List entitlements in a project",
		RunE: func(c *cobra.Command, args []string) error {
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

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Lookup Key", "Display Name", "State", "Created"})
				for _, e := range resp.Items {
					t.AppendRow(table.Row{e.ID, e.LookupKey, e.DisplayName, e.State, output.FormatTimestamp(e.CreatedAt)})
				}
			})
			return nil
		},
	}
}

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <entitlement-id>",
		Short: "Get an entitlement by ID",
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

	cmd.Flags().StringVar(&lookupKey, "lookup-key", "", "lookup key identifier (required)")
	cmd.Flags().StringVar(&displayName, "display-name", "", "display name (required)")
	cmd.MarkFlagRequired("lookup-key")
	cmd.MarkFlagRequired("display-name")
	return cmd
}

func newUpdateCmd(projectID, outputFormat *string) *cobra.Command {
	var displayName string

	cmd := &cobra.Command{
		Use:   "update <entitlement-id>",
		Short: "Update an entitlement",
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
	cmd.MarkFlagRequired("display-name")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <entitlement-id>",
		Short: "Delete an entitlement",
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
	return &cobra.Command{
		Use:   "products <entitlement-id>",
		Short: "List products attached to an entitlement",
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

			data, err := client.Get(fmt.Sprintf("/projects/%s/entitlements/%s/products", url.PathEscape(pid), url.PathEscape(args[0])), nil)
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
			return nil
		},
	}
}

func newAttachCmd(projectID *string) *cobra.Command {
	var (
		entitlementID string
		productIDs    []string
	)

	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach products to an entitlement",
		Long: `Attach one or more products to an entitlement.

Examples:
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1a2b3c
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1,prod2`,
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
	cmd.MarkFlagRequired("entitlement-id")
	cmd.MarkFlagRequired("product-id")
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
	cmd.MarkFlagRequired("entitlement-id")
	cmd.MarkFlagRequired("product-id")
	return cmd
}
