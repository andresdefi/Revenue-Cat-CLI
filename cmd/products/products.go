package products

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
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var appID string

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

			query := url.Values{}
			if appID != "" {
				query.Set("app_id", appID)
			}

			data, err := client.Get(fmt.Sprintf("/projects/%s/products", url.PathEscape(pid)), query)
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
			return nil
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "filter by app ID")
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

	cmd.Flags().StringVar(&storeIdentifier, "store-id", "", "store product identifier (required)")
	cmd.Flags().StringVar(&appID, "app-id", "", "app ID (required)")
	cmd.Flags().StringVar(&productType, "type", "", "product type: subscription, one_time, consumable, non_consumable (required)")
	cmd.Flags().StringVar(&displayName, "display-name", "", "display name")
	cmd.MarkFlagRequired("store-id")
	cmd.MarkFlagRequired("app-id")
	cmd.MarkFlagRequired("type")
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
