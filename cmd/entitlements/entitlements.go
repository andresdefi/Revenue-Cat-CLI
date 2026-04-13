package entitlements

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewEntitlementsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "entitlements",
		Aliases: []string{"entitlement", "ent"},
		Short:   "Manage entitlements",
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
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
					t.AppendRow(table.Row{
						e.ID,
						e.LookupKey,
						e.DisplayName,
						e.State,
						output.FormatTimestamp(e.CreatedAt),
					})
				}
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

	createCmd := &cobra.Command{
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

			body := map[string]any{
				"lookup_key":   lookupKey,
				"display_name": displayName,
			}

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

	createCmd.Flags().StringVar(&lookupKey, "lookup-key", "", "lookup key identifier (required)")
	createCmd.Flags().StringVar(&displayName, "display-name", "", "display name (required)")
	createCmd.MarkFlagRequired("lookup-key")
	createCmd.MarkFlagRequired("display-name")
	return createCmd
}

func newAttachCmd(projectID *string) *cobra.Command {
	var (
		entitlementID string
		productIDs    []string
	)

	attachCmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach products to an entitlement",
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}

			client, err := api.NewClient()
			if err != nil {
				return err
			}

			body := map[string]any{
				"product_ids": productIDs,
			}

			_, err = client.Post(
				fmt.Sprintf("/projects/%s/entitlements/%s/actions/attach_products",
					url.PathEscape(pid), url.PathEscape(entitlementID)),
				body,
			)
			if err != nil {
				return err
			}

			output.Success("Attached %d product(s) to entitlement %s", len(productIDs), entitlementID)
			return nil
		},
	}

	attachCmd.Flags().StringVar(&entitlementID, "entitlement-id", "", "entitlement ID (required)")
	attachCmd.Flags().StringSliceVar(&productIDs, "product-id", nil, "product ID(s) to attach (required, repeatable)")
	attachCmd.MarkFlagRequired("entitlement-id")
	attachCmd.MarkFlagRequired("product-id")
	return attachCmd
}

func newDetachCmd(projectID *string) *cobra.Command {
	var (
		entitlementID string
		productIDs    []string
	)

	detachCmd := &cobra.Command{
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

			body := map[string]any{
				"product_ids": productIDs,
			}

			_, err = client.Post(
				fmt.Sprintf("/projects/%s/entitlements/%s/actions/detach_products",
					url.PathEscape(pid), url.PathEscape(entitlementID)),
				body,
			)
			if err != nil {
				return err
			}

			output.Success("Detached %d product(s) from entitlement %s", len(productIDs), entitlementID)
			return nil
		},
	}

	detachCmd.Flags().StringVar(&entitlementID, "entitlement-id", "", "entitlement ID (required)")
	detachCmd.Flags().StringSliceVar(&productIDs, "product-id", nil, "product ID(s) to detach (required, repeatable)")
	detachCmd.MarkFlagRequired("entitlement-id")
	detachCmd.MarkFlagRequired("product-id")
	return detachCmd
}
