package products

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

func NewProductsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "products",
		Aliases: []string{"product", "prod"},
		Short:   "Manage products",
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var appID string

	listCmd := &cobra.Command{
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
					t.AppendRow(table.Row{
						p.ID,
						p.StoreIdentifier,
						p.Type,
						p.State,
						output.Deref(p.DisplayName, "-"),
						output.FormatTimestamp(p.CreatedAt),
					})
				}
			})
			return nil
		},
	}

	listCmd.Flags().StringVar(&appID, "app-id", "", "filter by app ID")
	return listCmd
}

func newCreateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		storeIdentifier string
		appID           string
		productType     string
		displayName     string
	)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new product",
		Long: `Create a new product in a RevenueCat project.

The store identifier must match the product ID configured in the app store
(App Store Connect, Google Play Console, etc).`,
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

	createCmd.Flags().StringVar(&storeIdentifier, "store-id", "", "store product identifier (required)")
	createCmd.Flags().StringVar(&appID, "app-id", "", "app ID (required)")
	createCmd.Flags().StringVar(&productType, "type", "", "product type: subscription, one_time, consumable, non_consumable (required)")
	createCmd.Flags().StringVar(&displayName, "display-name", "", "display name")
	createCmd.MarkFlagRequired("store-id")
	createCmd.MarkFlagRequired("app-id")
	createCmd.MarkFlagRequired("type")

	return createCmd
}
