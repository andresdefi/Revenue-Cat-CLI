package purchases

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

func NewPurchasesCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "purchases",
		Aliases: []string{"purchase"},
		Short:   "Search and manage one-time purchases",
	}
	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newGetCmd(projectID, outputFormat))
	root.AddCommand(newEntitlementsCmd(projectID, outputFormat))
	root.AddCommand(newRefundCmd(projectID))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		storePurchaseID string
		fetchAll        bool
		limit           int
	)
	cmd := &cobra.Command{
		Use: "list", Short: "Search purchases by store purchase identifier",
		Example: `  # Search purchases by store purchase identifier
  rc purchases list --store-purchase-id 100001234567890

  # List with JSON output
  rc purchases list --store-purchase-id 100001234567890 -o json

  # Fetch all pages
  rc purchases list --store-purchase-id 100001234567890 --all`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/projects/%s/purchases", url.PathEscape(pid))
			query := url.Values{"store_purchase_identifier": []string{storePurchaseID}}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}
			if fetchAll {
				items, err := api.PaginateAll[api.Purchase](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Customer", "Product", "Status", "Qty", "Store", "Purchased"})
					for _, p := range items {
						t.AppendRow(table.Row{p.ID, p.CustomerID, p.ProductID, p.Status, p.Quantity, p.Store, output.FormatTimestamp(p.PurchasedAt)})
					}
					t.AppendFooter(table.Row{"", "", "", "", "", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}
			data, err := client.Get(path, query)
			if err != nil {
				return err
			}
			var resp api.ListResponse[api.Purchase]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Customer", "Product", "Status", "Qty", "Store", "Purchased"})
				for _, p := range resp.Items {
					t.AppendRow(table.Row{p.ID, p.CustomerID, p.ProductID, p.Status, p.Quantity, p.Store, output.FormatTimestamp(p.PurchasedAt)})
				}
			})
			if resp.NextPage != nil {
				output.Warn("More results available (use --all for more)")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&storePurchaseID, "store-purchase-id", "", "store purchase identifier to search for (required)")
	cmd.Flags().BoolVar(&fetchAll, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&limit, "limit", 0, "max items per page")
	cmdutil.MustMarkFlagRequired(cmd, "store-purchase-id")
	return cmd
}

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use: "get <purchase-id>", Short: "Get a purchase by ID",
		Example: `  # Get purchase details
  rc purchases get purch1a2b3c4d5

  # Get as JSON
  rc purchases get purch1a2b3c4d5 -o json`,
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
			data, err := client.Get(fmt.Sprintf("/projects/%s/purchases/%s", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			var purchase api.Purchase
			if err := json.Unmarshal(data, &purchase); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, purchase, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", purchase.ID},
					{"Customer", purchase.CustomerID},
					{"Product", purchase.ProductID},
					{"Status", purchase.Status},
					{"Quantity", purchase.Quantity},
					{"Store", purchase.Store},
					{"Environment", purchase.Environment},
					{"Ownership", purchase.Ownership},
					{"Country", output.Deref(purchase.Country, "-")},
					{"Purchased", output.FormatTimestamp(purchase.PurchasedAt)},
				})
				if purchase.RevenueInUSD != nil {
					t.AppendSeparator()
					t.AppendRow(table.Row{"Revenue (USD)", fmt.Sprintf("$%.2f gross, $%.2f proceeds", purchase.RevenueInUSD.Gross, purchase.RevenueInUSD.Proceeds)})
				}
			})
			return nil
		},
	}
}

func newEntitlementsCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use: "entitlements <purchase-id>", Short: "List entitlements for a purchase",
		Example: `  # List entitlements for a purchase
  rc purchases entitlements purch1a2b3c4d5`,
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
			data, err := client.Get(fmt.Sprintf("/projects/%s/purchases/%s/entitlements", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			var resp api.ListResponse[api.Entitlement]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Lookup Key", "Display Name"})
				for _, e := range resp.Items {
					t.AppendRow(table.Row{e.ID, e.LookupKey, e.DisplayName})
				}
			})
			return nil
		},
	}
}

func newRefundCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use: "refund <purchase-id>", Short: "Refund a purchase",
		Example: `  # Refund a purchase
  rc purchases refund purch1a2b3c4d5`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.ConfirmDestructive("Refund", "purchase", args[0]); err != nil {
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
			_, err = client.Post(fmt.Sprintf("/projects/%s/purchases/%s/actions/refund", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Purchase %s refunded", args[0])
			return nil
		},
	}
}
