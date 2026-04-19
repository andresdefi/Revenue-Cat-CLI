package subscriptions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/completions"
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
  rc subscriptions list --store-subscription-id 100001234567890
  rc subscriptions get sub1ab2c3d4e5
  rc subscriptions transactions sub1ab2c3d4e5
  rc subscriptions cancel sub1ab2c3d4e5
  rc subscriptions refund sub1ab2c3d4e5`,
	}

	c := completions.SubscriptionIDs(projectID)
	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(completions.WithCompletion(newGetCmd(projectID, outputFormat), c))
	root.AddCommand(completions.WithCompletion(newTransactionsCmd(projectID, outputFormat), c))
	root.AddCommand(completions.WithCompletion(newEntitlementsCmd(projectID, outputFormat), c))
	root.AddCommand(completions.WithCompletion(newCancelCmd(projectID), c))
	root.AddCommand(completions.WithCompletion(newRefundCmd(projectID), c))
	root.AddCommand(completions.WithCompletion(newRefundTransactionCmd(projectID), c))
	root.AddCommand(completions.WithCompletion(newManagementURLCmd(projectID, outputFormat), c))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		storeSubscriptionID string
		fetchAll            bool
		limit               int
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Search subscriptions by store subscription identifier",
		Example: `  # Search subscriptions by store subscription identifier
  rc subscriptions list --store-subscription-id 100001234567890

  # Search subscriptions for a specific project as JSON
  rc subscriptions list --store-subscription-id 100001234567890 --project proj1a2b3c4d5 --output json

  # Use a production profile
  rc subscriptions list --store-subscription-id 100001234567890 --profile production

  # Extract active subscription IDs
  rc subscriptions list --store-subscription-id 100001234567890 --output json | jq -r '.items[] | select(.status == "active") | .id'

  # Find a subscription, then inspect transactions
  rc subscriptions list --store-subscription-id 100001234567890 --output json | jq -r '.items[0].id'
  rc subscriptions transactions sub1ab2c3d4e5

  # Fetch every page
  rc subscriptions list --store-subscription-id 100001234567890 --all --limit 100`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/projects/%s/subscriptions", url.PathEscape(pid))
			query := url.Values{"store_subscription_identifier": []string{storeSubscriptionID}}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}
			if fetchAll {
				items, err := api.PaginateAll[api.Subscription](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Customer", "Product", "Status", "Renewal", "Store", "Env"})
					for _, s := range items {
						t.AppendRow(table.Row{s.ID, s.CustomerID, output.Deref(s.ProductID, "promo"), s.Status, s.AutoRenewalStatus, s.Store, s.Environment})
					}
					t.AppendFooter(table.Row{"", "", "", "", "", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}
			data, err := client.Get(path, query)
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
			if resp.NextPage != nil {
				output.Warn("More results available (use --all for more)")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&storeSubscriptionID, "store-subscription-id", "", "store subscription identifier to search for (required)")
	cmd.Flags().BoolVar(&fetchAll, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&limit, "limit", 0, "max items per page")
	cmdutil.MustMarkFlagRequired(cmd, "store-subscription-id")
	return cmd
}

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		watch    bool
		interval time.Duration
	)
	cmd := &cobra.Command{
		Use:   "get <subscription-id>",
		Short: "Get a subscription by ID",
		Example: `  # Get subscription details
  rc subscriptions get sub1ab2c3d4e5

  # Get as JSON
  rc subscriptions get sub1ab2c3d4e5 --output json

  # Use a production profile
  rc subscriptions get sub1ab2c3d4e5 --profile production

  # Extract the authenticated management URL
  rc subscriptions management-url sub1ab2c3d4e5 --output json | jq -r '.management_url'

  # Inspect a subscription, then list transactions
  rc subscriptions get sub1ab2c3d4e5
  rc subscriptions transactions sub1ab2c3d4e5 --output json

  # Watch for changes
  rc subscriptions get sub1ab2c3d4e5 --watch --interval 10s`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			run := func(_ context.Context) error {
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
			}
			if watch {
				return cmdutil.Watch(c.Context(), interval, run)
			}
			return run(c.Context())
		},
	}
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "continuously refresh")
	cmd.Flags().DurationVar(&interval, "interval", cmdutil.DefaultWatchInterval, "refresh interval for --watch")
	return cmd
}

func newTransactionsCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		fetchAll bool
		limit    int
	)
	cmd := &cobra.Command{
		Use: "transactions <subscription-id>", Short: "List transactions for a subscription",
		Example: `  # List transactions
  rc subscriptions transactions sub1ab2c3d4e5

  # Fetch all pages
  rc subscriptions transactions sub1ab2c3d4e5 --all`,
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
			path := fmt.Sprintf("/projects/%s/subscriptions/%s/transactions", url.PathEscape(pid), url.PathEscape(args[0]))
			query := url.Values{}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}
			if fetchAll {
				items, err := api.PaginateAll[api.Transaction](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Store", "Purchased", "Revenue (USD)"})
					for _, tx := range items {
						rev := "-"
						if tx.RevenueInUSD != nil {
							rev = fmt.Sprintf("$%.2f", tx.RevenueInUSD.Gross)
						}
						t.AppendRow(table.Row{tx.ID, tx.Store, output.FormatTimestamp(tx.PurchasedAt), rev})
					}
					t.AppendFooter(table.Row{"", "", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}
			data, err := client.Get(path, query)
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

func newEntitlementsCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use: "entitlements <subscription-id>", Short: "List entitlements for a subscription",
		Example: `  # List entitlements for a subscription
  rc subscriptions entitlements sub1ab2c3d4e5`,
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
		Use: "cancel <subscription-id>", Short: "Cancel a subscription",
		Example: `  # Cancel a subscription
  rc subscriptions cancel sub1ab2c3d4e5`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.ConfirmDestructive("Cancel", "subscription", args[0]); err != nil {
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
			_, err = client.Post(fmt.Sprintf("/projects/%s/subscriptions/%s/actions/cancel", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Subscription %s canceled", args[0])
			output.Next("rc subscriptions get %s to verify", args[0])
			return nil
		},
	}
}

func newRefundCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use: "refund <subscription-id>", Short: "Refund a subscription",
		Example: `  # Refund a subscription
  rc subscriptions refund sub1ab2c3d4e5`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.ConfirmDestructive("Refund", "subscription", args[0]); err != nil {
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
		Use: "refund-transaction <subscription-id>", Short: "Refund a specific transaction within a subscription",
		Example: `  # Refund a specific transaction
  rc subscriptions refund-transaction sub1ab2c3d4e5 --transaction-id txn1a2b3c`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.ConfirmDestructive("Refund", "transaction", transactionID); err != nil {
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
	cmdutil.MustMarkFlagRequired(cmd, "transaction-id")
	return cmd
}

func newManagementURLCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use: "management-url <subscription-id>", Short: "Get authenticated management URL for a subscription",
		Example: `  # Get management URL
  rc subscriptions management-url sub1ab2c3d4e5`,
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
				t.AppendRow(table.Row{mgmt.ManagementURL})
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
