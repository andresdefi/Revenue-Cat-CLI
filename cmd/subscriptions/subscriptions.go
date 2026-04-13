package subscriptions

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

func NewSubscriptionsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "subscriptions",
		Aliases: []string{"subscription", "sub"},
		Short:   "Manage subscriptions",
		Long: `View and manage RevenueCat subscriptions.

Examples:
  rc subscriptions list
  rc subscriptions get sub1ab2c3d4e5
  rc subscriptions transactions sub1ab2c3d4e5
  rc subscriptions cancel sub1ab2c3d4e5
  rc subscriptions refund sub1ab2c3d4e5`,
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newGetCmd(projectID, outputFormat))
	root.AddCommand(newTransactionsCmd(projectID, outputFormat))
	root.AddCommand(newEntitlementsCmd(projectID, outputFormat))
	root.AddCommand(newCancelCmd(projectID))
	root.AddCommand(newRefundCmd(projectID))
	root.AddCommand(newRefundTransactionCmd(projectID))
	root.AddCommand(newManagementURLCmd(projectID, outputFormat))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List subscriptions in a project",
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			data, err := client.Get(fmt.Sprintf("/projects/%s/subscriptions", url.PathEscape(pid)), nil)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.Subscription]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Customer", "Product", "Status", "Renewal", "Store", "Env"})
				for _, s := range resp.Items {
					t.AppendRow(table.Row{s.ID, s.CustomerID, output.Deref(s.ProductID, "promo"), s.Status, s.AutoRenewalStatus, s.Store, s.Environment})
				}
			})
			return nil
		},
	}
}

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <subscription-id>",
		Short: "Get a subscription by ID",
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

			data, err := client.Get(fmt.Sprintf("/projects/%s/subscriptions/%s", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}

			var sub api.Subscription
			if err := json.Unmarshal(data, &sub); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, sub, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", sub.ID},
					{"Customer", sub.CustomerID},
					{"Product", output.Deref(sub.ProductID, "promotional")},
					{"Status", sub.Status},
					{"Auto Renewal", sub.AutoRenewalStatus},
					{"Gives Access", sub.GivesAccess},
					{"Store", sub.Store},
					{"Environment", sub.Environment},
					{"Ownership", sub.Ownership},
					{"Country", output.Deref(sub.Country, "-")},
					{"Starts At", output.FormatTimestamp(sub.StartsAt)},
					{"Period Start", output.FormatTimestamp(sub.CurrentPeriodStartsAt)},
					{"Period End", formatOptionalTimestamp(sub.CurrentPeriodEndsAt)},
					{"Ends At", formatOptionalTimestamp(sub.EndsAt)},
				})
				if sub.TotalRevenueInUSD != nil {
					t.AppendSeparator()
					t.AppendRow(table.Row{"Revenue (USD)", fmt.Sprintf("$%.2f gross, $%.2f proceeds", sub.TotalRevenueInUSD.Gross, sub.TotalRevenueInUSD.Proceeds)})
				}
			})
			return nil
		},
	}
}

func newTransactionsCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "transactions <subscription-id>",
		Short: "List transactions for a subscription",
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

			data, err := client.Get(fmt.Sprintf("/projects/%s/subscriptions/%s/transactions", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.Transaction]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Store", "Purchased", "Revenue (USD)"})
				for _, tx := range resp.Items {
					rev := "-"
					if tx.RevenueInUSD != nil {
						rev = fmt.Sprintf("$%.2f", tx.RevenueInUSD.Gross)
					}
					t.AppendRow(table.Row{tx.ID, tx.Store, output.FormatTimestamp(tx.PurchasedAt), rev})
				}
			})
			return nil
		},
	}
}

func newEntitlementsCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "entitlements <subscription-id>",
		Short: "List entitlements for a subscription",
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

			data, err := client.Get(fmt.Sprintf("/projects/%s/subscriptions/%s/entitlements", url.PathEscape(pid), url.PathEscape(args[0])), nil)
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

func newCancelCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <subscription-id>",
		Short: "Cancel a subscription",
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
			_, err = client.Post(fmt.Sprintf("/projects/%s/subscriptions/%s/actions/cancel", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Subscription %s cancelled", args[0])
			return nil
		},
	}
}

func newRefundCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "refund <subscription-id>",
		Short: "Refund a subscription",
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
			_, err = client.Post(fmt.Sprintf("/projects/%s/subscriptions/%s/actions/refund", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Subscription %s refunded", args[0])
			return nil
		},
	}
}

func newRefundTransactionCmd(projectID *string) *cobra.Command {
	var transactionID string

	cmd := &cobra.Command{
		Use:   "refund-transaction <subscription-id>",
		Short: "Refund a specific transaction within a subscription",
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
			_, err = client.Post(
				fmt.Sprintf("/projects/%s/subscriptions/%s/transactions/%s/actions/refund",
					url.PathEscape(pid), url.PathEscape(args[0]), url.PathEscape(transactionID)),
				nil,
			)
			if err != nil {
				return err
			}
			output.Success("Transaction %s refunded", transactionID)
			return nil
		},
	}

	cmd.Flags().StringVar(&transactionID, "transaction-id", "", "transaction ID to refund (required)")
	cmd.MarkFlagRequired("transaction-id")
	return cmd
}

func newManagementURLCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "management-url <subscription-id>",
		Short: "Get authenticated management URL for a subscription",
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

			data, err := client.Get(
				fmt.Sprintf("/projects/%s/subscriptions/%s/authenticated_management_url", url.PathEscape(pid), url.PathEscape(args[0])), nil,
			)
			if err != nil {
				return err
			}

			var mgmt api.ManagementURL
			if err := json.Unmarshal(data, &mgmt); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, mgmt, func(t table.Writer) {
				t.AppendHeader(table.Row{"Management URL"})
				t.AppendRow(table.Row{mgmt.URL})
			})
			return nil
		},
	}
}

func formatOptionalTimestamp(ms *int64) string {
	if ms == nil {
		return "-"
	}
	return output.FormatTimestamp(*ms)
}
