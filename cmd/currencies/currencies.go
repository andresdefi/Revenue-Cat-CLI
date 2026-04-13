package currencies

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

func NewCurrenciesCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "currencies",
		Aliases: []string{"currency", "vc"},
		Short:   "Manage virtual currencies",
		Long: `Manage virtual currencies in a RevenueCat project.

Virtual currencies let you define in-app currency systems (coins, gems, etc.)
that can be granted to customers.

Examples:
  rc currencies list
  rc currencies get COINS
  rc currencies create --code COINS --name "Gold Coins"
  rc currencies balance --customer-id user-123
  rc currencies credit --customer-id user-123 --code COINS --amount 100`,
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newGetCmd(projectID, outputFormat))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	root.AddCommand(newUpdateCmd(projectID, outputFormat))
	root.AddCommand(newDeleteCmd(projectID))
	root.AddCommand(newArchiveCmd(projectID))
	root.AddCommand(newUnarchiveCmd(projectID))
	root.AddCommand(newBalanceCmd(projectID, outputFormat))
	root.AddCommand(newCreditCmd(projectID))
	root.AddCommand(newUpdateBalanceCmd(projectID))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List virtual currencies",
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			data, err := client.Get(fmt.Sprintf("/projects/%s/virtual_currencies", url.PathEscape(pid)), nil)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.VirtualCurrency]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"Code", "Name", "State", "Created"})
				for _, vc := range resp.Items {
					t.AppendRow(table.Row{vc.Code, vc.Name, vc.State, output.FormatTimestamp(vc.CreatedAt)})
				}
			})
			return nil
		},
	}
}

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <currency-code>",
		Short: "Get a virtual currency by code",
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

			data, err := client.Get(fmt.Sprintf("/projects/%s/virtual_currencies/%s", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}

			var vc api.VirtualCurrency
			if err := json.Unmarshal(data, &vc); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, vc, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"Code", vc.Code},
					{"Name", vc.Name},
					{"State", vc.State},
					{"Created", output.FormatTimestamp(vc.CreatedAt)},
				})
			})
			return nil
		},
	}
}

func newCreateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		code string
		name string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a virtual currency",
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
				fmt.Sprintf("/projects/%s/virtual_currencies", url.PathEscape(pid)),
				map[string]any{"code": code, "name": name},
			)
			if err != nil {
				return err
			}

			var vc api.VirtualCurrency
			if err := json.Unmarshal(data, &vc); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, vc, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"Code", vc.Code},
					{"Name", vc.Name},
					{"State", vc.State},
				})
			})
			output.Success("Virtual currency created")
			return nil
		},
	}

	cmd.Flags().StringVar(&code, "code", "", "currency code, e.g. COINS (required)")
	cmd.Flags().StringVar(&name, "name", "", "display name (required)")
	cmd.MarkFlagRequired("code")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newUpdateCmd(projectID, outputFormat *string) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "update <currency-code>",
		Short: "Update a virtual currency",
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

			data, err := client.Post(
				fmt.Sprintf("/projects/%s/virtual_currencies/%s", url.PathEscape(pid), url.PathEscape(args[0])),
				map[string]any{"name": name},
			)
			if err != nil {
				return err
			}

			var vc api.VirtualCurrency
			if err := json.Unmarshal(data, &vc); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, vc, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"Code", vc.Code},
					{"Name", vc.Name},
				})
			})
			output.Success("Virtual currency updated")
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new display name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <currency-code>",
		Short: "Delete a virtual currency",
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
			_, err = client.Delete(fmt.Sprintf("/projects/%s/virtual_currencies/%s", url.PathEscape(pid), url.PathEscape(args[0])))
			if err != nil {
				return err
			}
			output.Success("Virtual currency %s deleted", args[0])
			return nil
		},
	}
}

func newArchiveCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <currency-code>",
		Short: "Archive a virtual currency",
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
			_, err = client.Post(fmt.Sprintf("/projects/%s/virtual_currencies/%s/actions/archive", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Virtual currency %s archived", args[0])
			return nil
		},
	}
}

func newUnarchiveCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "unarchive <currency-code>",
		Short: "Unarchive a virtual currency",
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
			_, err = client.Post(fmt.Sprintf("/projects/%s/virtual_currencies/%s/actions/unarchive", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Virtual currency %s unarchived", args[0])
			return nil
		},
	}
}

func newBalanceCmd(projectID, outputFormat *string) *cobra.Command {
	var customerID string

	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Show a customer's virtual currency balances",
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
				fmt.Sprintf("/projects/%s/customers/%s/virtual_currencies", url.PathEscape(pid), url.PathEscape(customerID)), nil,
			)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.VCBalance]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"Currency", "Balance"})
				for _, b := range resp.Items {
					t.AppendRow(table.Row{b.CurrencyCode, b.Balance})
				}
			})
			return nil
		},
	}

	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.MarkFlagRequired("customer-id")
	return cmd
}

func newCreditCmd(projectID *string) *cobra.Command {
	var (
		customerID string
		code       string
		amount     int64
	)

	cmd := &cobra.Command{
		Use:   "credit",
		Short: "Create a virtual currency transaction (credit/debit)",
		Long: `Create a virtual currency transaction for a customer.
Use positive amounts for credits and negative for debits.

Examples:
  rc currencies credit --customer-id user-123 --code COINS --amount 100
  rc currencies credit --customer-id user-123 --code COINS --amount -50`,
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
				fmt.Sprintf("/projects/%s/customers/%s/virtual_currencies/transactions", url.PathEscape(pid), url.PathEscape(customerID)),
				map[string]any{"currency_code": code, "amount": amount},
			)
			if err != nil {
				return err
			}
			output.Success("Transaction created: %+d %s for customer %s", amount, code, customerID)
			return nil
		},
	}

	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringVar(&code, "code", "", "currency code (required)")
	cmd.Flags().Int64Var(&amount, "amount", 0, "amount (positive=credit, negative=debit) (required)")
	cmd.MarkFlagRequired("customer-id")
	cmd.MarkFlagRequired("code")
	cmd.MarkFlagRequired("amount")
	return cmd
}

func newUpdateBalanceCmd(projectID *string) *cobra.Command {
	var (
		customerID string
		code       string
		balance    int64
	)

	cmd := &cobra.Command{
		Use:   "set-balance",
		Short: "Set a customer's virtual currency balance directly",
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
				fmt.Sprintf("/projects/%s/customers/%s/virtual_currencies/update_balance", url.PathEscape(pid), url.PathEscape(customerID)),
				map[string]any{"currency_code": code, "balance": balance},
			)
			if err != nil {
				return err
			}
			output.Success("Balance set to %d %s for customer %s", balance, code, customerID)
			return nil
		},
	}

	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringVar(&code, "code", "", "currency code (required)")
	cmd.Flags().Int64Var(&balance, "balance", 0, "new balance value (required)")
	cmd.MarkFlagRequired("customer-id")
	cmd.MarkFlagRequired("code")
	cmd.MarkFlagRequired("balance")
	return cmd
}
