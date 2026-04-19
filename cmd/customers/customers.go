package customers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewCustomersCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "customers",
		Aliases: []string{"customer", "cust"},
		Short:   "Manage customers and their entitlements",
		Long: `Look up, create, and manage RevenueCat customers.

Customers are identified by their app user ID. You can look up customer info,
check active entitlements, grant/revoke access, and manage subscriptions.

Examples:
  rc customers list
  rc customers lookup user-123
  rc customers diagnose user-123
  rc customers entitlements user-123
  rc customers subscriptions user-123
  rc customers grant --customer-id user-123 --entitlement-id entla1b2c3 --expires-at 1735689600000
  rc customers delete user-123`,
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newLookupCmd(projectID, outputFormat))
	root.AddCommand(newDiagnoseCmd(projectID, outputFormat))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	root.AddCommand(newDeleteCmd(projectID))
	root.AddCommand(newEntitlementsCmd(projectID, outputFormat))
	root.AddCommand(newSubscriptionsCmd(projectID, outputFormat))
	root.AddCommand(newPurchasesCmd(projectID, outputFormat))
	root.AddCommand(newAliasesCmd(projectID, outputFormat))
	root.AddCommand(newAttributesCmd(projectID, outputFormat))
	root.AddCommand(newSetAttributesCmd(projectID))
	root.AddCommand(newGrantCmd(projectID))
	root.AddCommand(newRevokeCmd(projectID))
	root.AddCommand(newAssignOfferingCmd(projectID))
	root.AddCommand(newTransferCmd(projectID))
	root.AddCommand(newRestorePurchaseCmd(projectID))
	root.AddCommand(newInvoicesCmd(projectID, outputFormat))
	root.AddCommand(newInvoiceFileCmd(projectID))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		search   string
		fetchAll bool
		limit    int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customers in a project",
		Example: `  # List customers
  rc customers list

  # Search by email
  rc customers list --search user@example.com

  # Fetch all pages
  rc customers list --all`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			path := fmt.Sprintf("/projects/%s/customers", url.PathEscape(pid))
			query := url.Values{}
			if search != "" {
				query.Set("search", search)
			}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}

			if fetchAll {
				items, err := api.PaginateAll[api.Customer](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Platform", "Country", "Entitlements", "First Seen"})
					for _, cu := range items {
						t.AppendRow(table.Row{
							cu.ID,
							output.Deref(cu.LastSeenPlatform, "-"),
							output.Deref(cu.LastSeenCountry, "-"),
							len(cu.ActiveEntitlements.Items),
							output.FormatTimestamp(cu.FirstSeenAt),
						})
					}
					t.AppendFooter(table.Row{"", "", "", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}

			data, err := client.Get(path, query)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.Customer]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Platform", "Country", "Entitlements", "First Seen"})
				for _, cu := range resp.Items {
					t.AppendRow(table.Row{
						cu.ID,
						output.Deref(cu.LastSeenPlatform, "-"),
						output.Deref(cu.LastSeenCountry, "-"),
						len(cu.ActiveEntitlements.Items),
						output.FormatTimestamp(cu.FirstSeenAt),
					})
				}
			})
			if resp.NextPage != nil {
				output.Warn("More results available (use --all for more)")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "search by email (exact match)")
	cmd.Flags().BoolVar(&fetchAll, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&limit, "limit", 0, "max items per page")
	return cmd
}

func newLookupCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		watch    bool
		interval time.Duration
	)

	cmd := &cobra.Command{
		Use:   "lookup <customer-id>",
		Short: "Look up a customer by ID",
		Example: `  # Look up a customer
  rc customers lookup user-123

  # Look up a customer as JSON
  rc customers lookup user-123 --output json

  # Use a production profile
  rc customers lookup user-123 --profile production

  # Extract active entitlement IDs
  rc customers lookup user-123 --output json | jq -r '.active_entitlements.items[].entitlement_id'

  # Look up a customer, then inspect their subscriptions
  rc customers lookup user-123
  rc customers subscriptions user-123 --output json

  # Watch for changes
  rc customers lookup user-123 --watch --interval 10s`,
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

				data, err := client.Get(
					fmt.Sprintf("/projects/%s/customers/%s", url.PathEscape(pid), url.PathEscape(args[0])), nil,
				)
				if err != nil {
					return err
				}

				var customer api.Customer
				if err := json.Unmarshal(data, &customer); err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}

				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, customer, func(t table.Writer) {
					t.AppendHeader(table.Row{"Field", "Value"})
					t.AppendRows([]table.Row{
						{"ID", customer.ID},
						{"First Seen", output.FormatTimestamp(customer.FirstSeenAt)},
						{"Last Seen", formatOptionalTimestamp(customer.LastSeenAt)},
						{"Platform", output.Deref(customer.LastSeenPlatform, "-")},
						{"Country", output.Deref(customer.LastSeenCountry, "-")},
						{"App Version", output.Deref(customer.LastSeenAppVersion, "-")},
						{"Active Entitlements", len(customer.ActiveEntitlements.Items)},
					})
					if len(customer.ActiveEntitlements.Items) > 0 {
						t.AppendSeparator()
						for _, e := range customer.ActiveEntitlements.Items {
							expires := "never"
							if e.ExpiresAt != nil {
								expires = output.FormatTimestamp(*e.ExpiresAt)
							}
							t.AppendRow(table.Row{e.EntitlementID, "expires: " + expires})
						}
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

func newCreateCmd(projectID, outputFormat *string) *cobra.Command {
	var customerID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a customer",
		Long: `Create a customer. Required flags are prompted interactively when
running in a terminal and not provided on the command line.`,
		Example: `  # Create a customer
  rc customers create --id user-456

  # Create and output as JSON
  rc customers create --id user-456 -o json

  # Interactive mode (prompts for missing fields)
  rc customers create`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.PromptIfEmpty(&customerID, "Customer ID", "user-456"); err != nil {
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
			data, err := client.Post(
				fmt.Sprintf("/projects/%s/customers", url.PathEscape(pid)),
				map[string]any{"id": customerID},
			)
			if err != nil {
				return err
			}
			var customer api.Customer
			if err := json.Unmarshal(data, &customer); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, customer, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", customer.ID},
					{"First Seen", output.FormatTimestamp(customer.FirstSeenAt)},
				})
			})
			output.Success("Customer created")
			output.Next("rc customers lookup %s", customer.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&customerID, "id", "", "customer ID (required)")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use: "delete <customer-id>", Short: "Delete a customer",
		Example: `  # Delete a customer
  rc customers delete user-123`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.ConfirmDestructive("Delete", "customer", args[0]); err != nil {
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
			_, err = client.Delete(fmt.Sprintf("/projects/%s/customers/%s", url.PathEscape(pid), url.PathEscape(args[0])))
			if err != nil {
				return err
			}
			output.Success("Customer %s deleted", args[0])
			return nil
		},
	}
}

func newEntitlementsCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		watch    bool
		interval time.Duration
	)

	cmd := &cobra.Command{
		Use:   "entitlements <customer-id>",
		Short: "List active entitlements for a customer",
		Example: `  # List active entitlements
  rc customers entitlements user-123

  # List active entitlements as JSON
  rc customers entitlements user-123 --output json

  # Use a production profile
  rc customers entitlements user-123 --profile production

  # Extract entitlement IDs for scripting
  rc customers entitlements user-123 --output json | jq -r '.items[].entitlement_id'

  # Look up a customer, then list active entitlements
  rc customers lookup user-123
  rc customers entitlements user-123

  # Watch for entitlement changes
  rc customers entitlements user-123 --watch --interval 10s`,
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
				data, err := client.Get(
					fmt.Sprintf("/projects/%s/customers/%s/active_entitlements", url.PathEscape(pid), url.PathEscape(args[0])), nil,
				)
				if err != nil {
					return err
				}
				var resp api.ListResponse[api.ActiveEntitlement]
				if err := json.Unmarshal(data, &resp); err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, resp, func(t table.Writer) {
					t.AppendHeader(table.Row{"Entitlement ID", "Expires"})
					for _, e := range resp.Items {
						expires := "never"
						if e.ExpiresAt != nil {
							expires = output.FormatTimestamp(*e.ExpiresAt)
						}
						t.AppendRow(table.Row{e.EntitlementID, expires})
					}
					if len(resp.Items) == 0 {
						t.AppendRow(table.Row{"(no active entitlements)", ""})
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

func newSubscriptionsCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		fetchAll bool
		limit    int
		watch    bool
		interval time.Duration
	)
	cmd := &cobra.Command{
		Use: "subscriptions <customer-id>", Short: "List subscriptions for a customer",
		Example: `  # List customer subscriptions
  rc customers subscriptions user-123

  # Fetch all pages
  rc customers subscriptions user-123 --all

  # Watch for subscription changes
  rc customers subscriptions user-123 --watch --interval 10s`,
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
				path := fmt.Sprintf("/projects/%s/customers/%s/subscriptions", url.PathEscape(pid), url.PathEscape(args[0]))
				query := url.Values{}
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
						t.AppendHeader(table.Row{"ID", "Product", "Status", "Renewal", "Store", "Env"})
						for _, s := range items {
							t.AppendRow(table.Row{s.ID, output.Deref(s.ProductID, "promotional"), s.Status, s.AutoRenewalStatus, s.Store, s.Environment})
						}
						t.AppendFooter(table.Row{"", "", "", "", "", fmt.Sprintf("%d total", len(items))})
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
					t.AppendHeader(table.Row{"ID", "Product", "Status", "Renewal", "Store", "Env"})
					for _, s := range resp.Items {
						t.AppendRow(table.Row{s.ID, output.Deref(s.ProductID, "promotional"), s.Status, s.AutoRenewalStatus, s.Store, s.Environment})
					}
				})
				if resp.NextPage != nil {
					output.Warn("More results available (use --all for more)")
				}
				return nil
			}

			if watch {
				return cmdutil.Watch(c.Context(), interval, run)
			}
			return run(c.Context())
		},
	}
	cmd.Flags().BoolVar(&fetchAll, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&limit, "limit", 0, "max items per page")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "continuously refresh")
	cmd.Flags().DurationVar(&interval, "interval", cmdutil.DefaultWatchInterval, "refresh interval for --watch")
	return cmd
}

func newPurchasesCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		fetchAll bool
		limit    int
		watch    bool
		interval time.Duration
	)
	cmd := &cobra.Command{
		Use: "purchases <customer-id>", Short: "List purchases for a customer",
		Example: `  # List customer purchases
  rc customers purchases user-123

  # Fetch all pages
  rc customers purchases user-123 --all

  # Watch for purchase changes
  rc customers purchases user-123 --watch --interval 10s`,
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
				path := fmt.Sprintf("/projects/%s/customers/%s/purchases", url.PathEscape(pid), url.PathEscape(args[0]))
				query := url.Values{}
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
						t.AppendHeader(table.Row{"ID", "Product", "Status", "Qty", "Store", "Purchased"})
						for _, p := range items {
							t.AppendRow(table.Row{p.ID, p.ProductID, p.Status, p.Quantity, p.Store, output.FormatTimestamp(p.PurchasedAt)})
						}
						t.AppendFooter(table.Row{"", "", "", "", "", fmt.Sprintf("%d total", len(items))})
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
					t.AppendHeader(table.Row{"ID", "Product", "Status", "Qty", "Store", "Purchased"})
					for _, p := range resp.Items {
						t.AppendRow(table.Row{p.ID, p.ProductID, p.Status, p.Quantity, p.Store, output.FormatTimestamp(p.PurchasedAt)})
					}
				})
				if resp.NextPage != nil {
					output.Warn("More results available (use --all for more)")
				}
				return nil
			}

			if watch {
				return cmdutil.Watch(c.Context(), interval, run)
			}
			return run(c.Context())
		},
	}
	cmd.Flags().BoolVar(&fetchAll, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&limit, "limit", 0, "max items per page")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "continuously refresh")
	cmd.Flags().DurationVar(&interval, "interval", cmdutil.DefaultWatchInterval, "refresh interval for --watch")
	return cmd
}

func newAliasesCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		watch    bool
		interval time.Duration
	)

	cmd := &cobra.Command{
		Use: "aliases <customer-id>", Short: "List aliases for a customer",
		Example: `  # List aliases
  rc customers aliases user-123

  # Watch for alias changes
  rc customers aliases user-123 --watch --interval 10s`,
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
				data, err := client.Get(fmt.Sprintf("/projects/%s/customers/%s/aliases", url.PathEscape(pid), url.PathEscape(args[0])), nil)
				if err != nil {
					return err
				}
				var resp api.ListResponse[api.CustomerAlias]
				if err := json.Unmarshal(data, &resp); err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, resp, func(t table.Writer) {
					t.AppendHeader(table.Row{"Alias ID"})
					for _, a := range resp.Items {
						t.AppendRow(table.Row{a.ID})
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

func newAttributesCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		watch    bool
		interval time.Duration
	)

	cmd := &cobra.Command{
		Use: "attributes <customer-id>", Short: "List attributes for a customer",
		Example: `  # List customer attributes
  rc customers attributes user-123

  # Watch for attribute changes
  rc customers attributes user-123 --watch --interval 10s`,
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
				data, err := client.Get(fmt.Sprintf("/projects/%s/customers/%s/attributes", url.PathEscape(pid), url.PathEscape(args[0])), nil)
				if err != nil {
					return err
				}
				var resp api.ListResponse[api.CustomerAttribute]
				if err := json.Unmarshal(data, &resp); err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, resp, func(t table.Writer) {
					t.AppendHeader(table.Row{"Name", "Value"})
					for _, a := range resp.Items {
						t.AppendRow(table.Row{a.Name, a.Value})
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

func newSetAttributesCmd(projectID *string) *cobra.Command {
	var (
		customerID string
		attrs      []string
	)
	cmd := &cobra.Command{
		Use:   "set-attributes",
		Short: "Set attributes on a customer",
		Long:  `Set key=value attributes on a customer.`,
		Example: `  # Set a single attribute
  rc customers set-attributes --customer-id user-123 --attr 'plan=pro'

  # Set multiple attributes
  rc customers set-attributes --customer-id user-123 --attr '$email=user@example.com' --attr 'plan=pro'`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			var attrList []map[string]string
			for _, a := range attrs {
				for i, ch := range a {
					if ch == '=' {
						attrList = append(attrList, map[string]string{"name": a[:i], "value": a[i+1:]})
						break
					}
				}
			}
			_, err = client.Post(
				fmt.Sprintf("/projects/%s/customers/%s/attributes", url.PathEscape(pid), url.PathEscape(customerID)),
				map[string]any{"attributes": attrList},
			)
			if err != nil {
				return err
			}
			output.Success("Attributes set on customer %s", customerID)
			output.Next("rc customers attributes %s", customerID)
			return nil
		},
	}
	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringSliceVar(&attrs, "attr", nil, "attribute as key=value (required, repeatable)")
	cmdutil.MustMarkFlagRequired(cmd, "customer-id")
	cmdutil.MustMarkFlagRequired(cmd, "attr")
	return cmd
}

func newGrantCmd(projectID *string) *cobra.Command {
	var (
		customerID    string
		entitlementID string
		expiresAt     int64
		duration      string
		lifetime      bool
	)
	cmd := &cobra.Command{
		Use: "grant", Short: "Grant an entitlement to a customer",
		Example: `  # Grant an entitlement with expiration
  rc customers grant --customer-id user-123 --entitlement-id entla1b2c3 --expires-at 1735689600000

  # Grant access for 30 days
  rc customers grant --customer-id user-123 --entitlement-id entla1b2c3 --duration 30d

  # Grant long-lived access
  rc customers grant --customer-id user-123 --entitlement-id entla1b2c3 --lifetime`,
		RunE: func(c *cobra.Command, args []string) error {
			resolvedExpiresAt, err := resolveGrantExpiresAt(c, expiresAt, duration, lifetime, time.Now())
			if err != nil {
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
				fmt.Sprintf("/projects/%s/customers/%s/actions/grant_entitlement", url.PathEscape(pid), url.PathEscape(customerID)),
				map[string]any{"entitlement_id": entitlementID, "expires_at": resolvedExpiresAt},
			)
			if err != nil {
				return err
			}
			output.Success("Entitlement %s granted to customer %s", entitlementID, customerID)
			output.Next("rc customers entitlements %s", customerID)
			return nil
		},
	}
	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringVar(&entitlementID, "entitlement-id", "", "entitlement ID to grant (required)")
	cmd.Flags().Int64Var(&expiresAt, "expires-at", 0, "expiration timestamp in ms since epoch")
	cmd.Flags().StringVar(&duration, "duration", "", "grant duration (for example 12h, 30d, 2w)")
	cmd.Flags().BoolVar(&lifetime, "lifetime", false, "grant long-lived access using a far-future expiration")
	cmdutil.MustMarkFlagRequired(cmd, "customer-id")
	cmdutil.MustMarkFlagRequired(cmd, "entitlement-id")
	return cmd
}

const lifetimeGrantExpiresAt int64 = 253402300799999 // 9999-12-31T23:59:59.999Z

func resolveGrantExpiresAt(cmd *cobra.Command, expiresAt int64, duration string, lifetime bool, now time.Time) (int64, error) {
	choices := 0
	if cmd.Flags().Changed("expires-at") {
		choices++
	}
	if duration != "" {
		choices++
	}
	if lifetime {
		choices++
	}
	if choices == 0 {
		return 0, fmt.Errorf("one of --expires-at, --duration, or --lifetime is required")
	}
	if choices > 1 {
		return 0, fmt.Errorf("use only one of --expires-at, --duration, or --lifetime")
	}
	if lifetime {
		return lifetimeGrantExpiresAt, nil
	}
	if duration != "" {
		d, err := parseGrantDuration(duration)
		if err != nil {
			return 0, err
		}
		return now.Add(d).UnixMilli(), nil
	}
	if expiresAt <= 0 {
		return 0, fmt.Errorf("--expires-at must be a positive millisecond timestamp")
	}
	return expiresAt, nil
}

func parseGrantDuration(value string) (time.Duration, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, fmt.Errorf("--duration is required")
	}
	if strings.HasSuffix(trimmed, "d") || strings.HasSuffix(trimmed, "w") {
		multiplier := 24 * time.Hour
		if strings.HasSuffix(trimmed, "w") {
			multiplier = 7 * 24 * time.Hour
		}
		number := strings.TrimSpace(trimmed[:len(trimmed)-1])
		amount, err := strconv.ParseFloat(number, 64)
		if err != nil || amount <= 0 {
			return 0, fmt.Errorf("invalid --duration %q (expected a positive duration like 12h, 30d, or 2w)", value)
		}
		return time.Duration(amount * float64(multiplier)), nil
	}
	duration, err := time.ParseDuration(trimmed)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("invalid --duration %q (expected a positive duration like 12h, 30d, or 2w)", value)
	}
	return duration, nil
}

func newRevokeCmd(projectID *string) *cobra.Command {
	var (
		customerID    string
		entitlementID string
	)
	cmd := &cobra.Command{
		Use: "revoke", Short: "Revoke a granted entitlement from a customer",
		Example: `  # Revoke an entitlement
  rc customers revoke --customer-id user-123 --entitlement-id entla1b2c3`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.ConfirmDestructive("Revoke entitlement from", "customer", customerID); err != nil {
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
				fmt.Sprintf("/projects/%s/customers/%s/actions/revoke_granted_entitlement", url.PathEscape(pid), url.PathEscape(customerID)),
				map[string]any{"entitlement_id": entitlementID},
			)
			if err != nil {
				return err
			}
			output.Success("Entitlement %s revoked from customer %s", entitlementID, customerID)
			output.Next("rc customers entitlements %s", customerID)
			return nil
		},
	}
	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringVar(&entitlementID, "entitlement-id", "", "entitlement ID to revoke (required)")
	cmdutil.MustMarkFlagRequired(cmd, "customer-id")
	cmdutil.MustMarkFlagRequired(cmd, "entitlement-id")
	return cmd
}

func newAssignOfferingCmd(projectID *string) *cobra.Command {
	var (
		customerID string
		offeringID string
		clear      bool
	)
	cmd := &cobra.Command{
		Use: "assign-offering", Short: "Assign or clear an offering override for a customer",
		Example: `  # Assign an offering override
  rc customers assign-offering --customer-id user-123 --offering-id ofrnge1a2b3c

  # Clear the offering override
  rc customers assign-offering --customer-id user-123 --clear`,
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
			if clear {
				body["offering_id"] = nil
			} else {
				body["offering_id"] = offeringID
			}
			_, err = client.Post(
				fmt.Sprintf("/projects/%s/customers/%s/actions/assign_offering", url.PathEscape(pid), url.PathEscape(customerID)),
				body,
			)
			if err != nil {
				return err
			}
			if clear {
				output.Success("Offering override cleared for customer %s", customerID)
				output.Next("rc customers lookup %s", customerID)
			} else {
				output.Success("Offering %s assigned to customer %s", offeringID, customerID)
				output.Next("rc customers lookup %s", customerID)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringVar(&offeringID, "offering-id", "", "offering ID to assign")
	cmd.Flags().BoolVar(&clear, "clear", false, "clear the offering override")
	cmdutil.MustMarkFlagRequired(cmd, "customer-id")
	return cmd
}

func newTransferCmd(projectID *string) *cobra.Command {
	var (
		customerID       string
		targetCustomerID string
	)
	cmd := &cobra.Command{
		Use: "transfer", Short: "Transfer subscriptions/purchases to another customer",
		Example: `  # Transfer purchases between customers
  rc customers transfer --customer-id user-123 --target-id user-456`,
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
				fmt.Sprintf("/projects/%s/customers/%s/actions/transfer", url.PathEscape(pid), url.PathEscape(customerID)),
				map[string]any{"target_customer_id": targetCustomerID},
			)
			if err != nil {
				return err
			}
			output.Success("Transferred from %s to %s", customerID, targetCustomerID)
			output.Next("rc customers lookup %s", targetCustomerID)
			return nil
		},
	}
	cmd.Flags().StringVar(&customerID, "customer-id", "", "source customer ID (required)")
	cmd.Flags().StringVar(&targetCustomerID, "target-id", "", "target customer ID (required)")
	cmdutil.MustMarkFlagRequired(cmd, "customer-id")
	cmdutil.MustMarkFlagRequired(cmd, "target-id")
	return cmd
}

func newRestorePurchaseCmd(projectID *string) *cobra.Command {
	var (
		customerID string
		orderID    string
	)
	cmd := &cobra.Command{
		Use: "restore-purchase", Short: "Restore a Google Play purchase by order ID",
		Example: `  # Restore a Google Play purchase
  rc customers restore-purchase --customer-id user-123 --order-id GPA.1234-5678-9012`,
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
				fmt.Sprintf("/projects/%s/customers/%s/actions/restore_purchase_by_order_id", url.PathEscape(pid), url.PathEscape(customerID)),
				map[string]any{"order_id": orderID},
			)
			if err != nil {
				return err
			}
			output.Success("Purchase restored for customer %s (order: %s)", customerID, orderID)
			output.Next("rc customers purchases %s", customerID)
			return nil
		},
	}
	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringVar(&orderID, "order-id", "", "Google Play order ID (required)")
	cmdutil.MustMarkFlagRequired(cmd, "customer-id")
	cmdutil.MustMarkFlagRequired(cmd, "order-id")
	return cmd
}

func newInvoicesCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use: "invoices <customer-id>", Short: "List invoices for a customer",
		Example: `  # List invoices
  rc customers invoices user-123`,
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
			data, err := client.Get(fmt.Sprintf("/projects/%s/customers/%s/invoices", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			var resp api.ListResponse[api.Invoice]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Created"})
				for _, inv := range resp.Items {
					t.AppendRow(table.Row{inv.ID, output.FormatTimestamp(inv.CreatedAt)})
				}
			})
			return nil
		},
	}
}

func newInvoiceFileCmd(projectID *string) *cobra.Command {
	var (
		customerID string
		invoiceID  string
	)
	cmd := &cobra.Command{
		Use:   "invoice-file",
		Short: "Download an invoice file",
		Long: `Download an invoice file for a customer.

The file content is written to stdout. Redirect to save.`,
		Example: `  # Download an invoice to a file
  rc customers invoice-file --customer-id user-123 --invoice-id inv1a2b3c > invoice.pdf`,
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
				fmt.Sprintf("/projects/%s/customers/%s/invoices/%s/file",
					url.PathEscape(pid), url.PathEscape(customerID), url.PathEscape(invoiceID)),
				nil,
			)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(c.OutOrStdout(), string(data))
			return err
		},
	}
	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringVar(&invoiceID, "invoice-id", "", "invoice ID (required)")
	cmdutil.MustMarkFlagRequired(cmd, "customer-id")
	cmdutil.MustMarkFlagRequired(cmd, "invoice-id")
	return cmd
}

func formatOptionalTimestamp(ms *int64) string {
	if ms == nil {
		return "-"
	}
	return output.FormatTimestamp(*ms)
}
