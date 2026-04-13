package customers

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
  rc customers entitlements user-123
  rc customers subscriptions user-123
  rc customers grant --customer-id user-123 --entitlement-id entla1b2c3 --expires-at 1735689600000
  rc customers delete user-123`,
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newLookupCmd(projectID, outputFormat))
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
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var search string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List customers in a project",
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
			if search != "" {
				query.Set("search", search)
			}

			data, err := client.Get(fmt.Sprintf("/projects/%s/customers", url.PathEscape(pid)), query)
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
				for _, c := range resp.Items {
					t.AppendRow(table.Row{
						c.ID,
						output.Deref(c.LastSeenPlatform, "-"),
						output.Deref(c.LastSeenCountry, "-"),
						len(c.ActiveEntitlements.Items),
						output.FormatTimestamp(c.FirstSeenAt),
					})
				}
			})
			return nil
		},
	}

	cmd.Flags().StringVar(&search, "search", "", "search by email (exact match)")
	return cmd
}

func newLookupCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "lookup <customer-id>",
		Short: "Look up a customer by ID",
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
		},
	}
}

func newCreateCmd(projectID, outputFormat *string) *cobra.Command {
	var customerID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a customer",
		RunE: func(c *cobra.Command, args []string) error {
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
			return nil
		},
	}

	cmd.Flags().StringVar(&customerID, "id", "", "customer ID (required)")
	cmd.MarkFlagRequired("id")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <customer-id>",
		Short: "Delete a customer",
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
	return &cobra.Command{
		Use:   "entitlements <customer-id>",
		Short: "List active entitlements for a customer",
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
		},
	}
}

func newSubscriptionsCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "subscriptions <customer-id>",
		Short: "List subscriptions for a customer",
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
				fmt.Sprintf("/projects/%s/customers/%s/subscriptions", url.PathEscape(pid), url.PathEscape(args[0])), nil,
			)
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
			return nil
		},
	}
}

func newPurchasesCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "purchases <customer-id>",
		Short: "List purchases for a customer",
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
				fmt.Sprintf("/projects/%s/customers/%s/purchases", url.PathEscape(pid), url.PathEscape(args[0])), nil,
			)
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
			return nil
		},
	}
}

func newAliasesCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "aliases <customer-id>",
		Short: "List aliases for a customer",
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
				fmt.Sprintf("/projects/%s/customers/%s/aliases", url.PathEscape(pid), url.PathEscape(args[0])), nil,
			)
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
		},
	}
}

func newAttributesCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "attributes <customer-id>",
		Short: "List attributes for a customer",
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
				fmt.Sprintf("/projects/%s/customers/%s/attributes", url.PathEscape(pid), url.PathEscape(args[0])), nil,
			)
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
		},
	}
}

func newSetAttributesCmd(projectID *string) *cobra.Command {
	var (
		customerID string
		attrs      []string
	)

	cmd := &cobra.Command{
		Use:   "set-attributes",
		Short: "Set attributes on a customer",
		Long: `Set key=value attributes on a customer.

Examples:
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
			return nil
		},
	}

	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringSliceVar(&attrs, "attr", nil, "attribute as key=value (required, repeatable)")
	cmd.MarkFlagRequired("customer-id")
	cmd.MarkFlagRequired("attr")
	return cmd
}

func newGrantCmd(projectID *string) *cobra.Command {
	var (
		customerID    string
		entitlementID string
		expiresAt     int64
	)

	cmd := &cobra.Command{
		Use:   "grant",
		Short: "Grant an entitlement to a customer",
		Long: `Grant an entitlement to a customer (creates a promotional subscription).

Examples:
  rc customers grant --customer-id user-123 --entitlement-id entla1b2c3 --expires-at 1735689600000`,
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
				fmt.Sprintf("/projects/%s/customers/%s/actions/grant_entitlement", url.PathEscape(pid), url.PathEscape(customerID)),
				map[string]any{"entitlement_id": entitlementID, "expires_at": expiresAt},
			)
			if err != nil {
				return err
			}
			output.Success("Entitlement %s granted to customer %s", entitlementID, customerID)
			return nil
		},
	}

	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringVar(&entitlementID, "entitlement-id", "", "entitlement ID to grant (required)")
	cmd.Flags().Int64Var(&expiresAt, "expires-at", 0, "expiration timestamp in ms since epoch (required)")
	cmd.MarkFlagRequired("customer-id")
	cmd.MarkFlagRequired("entitlement-id")
	cmd.MarkFlagRequired("expires-at")
	return cmd
}

func newRevokeCmd(projectID *string) *cobra.Command {
	var (
		customerID    string
		entitlementID string
	)

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a granted entitlement from a customer",
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
				fmt.Sprintf("/projects/%s/customers/%s/actions/revoke_granted_entitlement", url.PathEscape(pid), url.PathEscape(customerID)),
				map[string]any{"entitlement_id": entitlementID},
			)
			if err != nil {
				return err
			}
			output.Success("Entitlement %s revoked from customer %s", entitlementID, customerID)
			return nil
		},
	}

	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringVar(&entitlementID, "entitlement-id", "", "entitlement ID to revoke (required)")
	cmd.MarkFlagRequired("customer-id")
	cmd.MarkFlagRequired("entitlement-id")
	return cmd
}

func newAssignOfferingCmd(projectID *string) *cobra.Command {
	var (
		customerID string
		offeringID string
		clear      bool
	)

	cmd := &cobra.Command{
		Use:   "assign-offering",
		Short: "Assign or clear an offering override for a customer",
		Long: `Assign a specific offering to a customer (override), or clear the override.

Examples:
  rc customers assign-offering --customer-id user-123 --offering-id ofrnge1a2b3c
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
			} else {
				output.Success("Offering %s assigned to customer %s", offeringID, customerID)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringVar(&offeringID, "offering-id", "", "offering ID to assign")
	cmd.Flags().BoolVar(&clear, "clear", false, "clear the offering override")
	cmd.MarkFlagRequired("customer-id")
	return cmd
}

func newTransferCmd(projectID *string) *cobra.Command {
	var (
		customerID       string
		targetCustomerID string
	)

	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Transfer subscriptions/purchases to another customer",
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
			return nil
		},
	}

	cmd.Flags().StringVar(&customerID, "customer-id", "", "source customer ID (required)")
	cmd.Flags().StringVar(&targetCustomerID, "target-id", "", "target customer ID (required)")
	cmd.MarkFlagRequired("customer-id")
	cmd.MarkFlagRequired("target-id")
	return cmd
}

func formatOptionalTimestamp(ms *int64) string {
	if ms == nil {
		return "-"
	}
	return output.FormatTimestamp(*ms)
}
