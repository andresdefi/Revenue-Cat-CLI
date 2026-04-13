package customers

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

func NewCustomersCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "customers",
		Aliases: []string{"customer", "cust"},
		Short:   "Look up customers and their entitlements",
	}

	root.AddCommand(newLookupCmd(projectID, outputFormat))
	root.AddCommand(newEntitlementsCmd(projectID, outputFormat))
	return root
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

			customerID := args[0]
			data, err := client.Get(
				fmt.Sprintf("/projects/%s/customers/%s", url.PathEscape(pid), url.PathEscape(customerID)),
				nil,
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
					t.AppendRow(table.Row{"", ""})
					t.AppendRow(table.Row{"ENTITLEMENTS", ""})
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

			customerID := args[0]
			data, err := client.Get(
				fmt.Sprintf("/projects/%s/customers/%s/active_entitlements",
					url.PathEscape(pid), url.PathEscape(customerID)),
				nil,
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

func formatOptionalTimestamp(ms *int64) string {
	if ms == nil {
		return "-"
	}
	return output.FormatTimestamp(*ms)
}
