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
	var (
		fetchAll bool
		limit    int
	)
	cmd := &cobra.Command{
		Use: "list", Short: "List virtual currencies",
		Example: `  # List virtual currencies
  rc currencies list

  # List with JSON output
  rc currencies list -o json`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/projects/%s/virtual_currencies", url.PathEscape(pid))
			query := url.Values{}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}
			if fetchAll {
				items, err := api.PaginateAll[api.VirtualCurrency](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"Code", "Name", "State", "Created"})
					for _, vc := range items {
						t.AppendRow(table.Row{vc.Code, vc.Name, vc.State, output.FormatTimestamp(vc.CreatedAt)})
					}
					t.AppendFooter(table.Row{"", "", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}
			data, err := client.Get(path, query)
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

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use: "get <currency-code>", Short: "Get a virtual currency by code",
		Example: `  # Get currency details
  rc currencies get COINS

  # Get as JSON
  rc currencies get COINS -o json`,
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
				t.AppendRows([]table.Row{{"Code", vc.Code}, {"Name", vc.Name}, {"State", vc.State}, {"Created", output.FormatTimestamp(vc.CreatedAt)}})
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
		Long: `Create a virtual currency. Required flags are prompted interactively
when running in a terminal and not provided on the command line.`,
		Example: `  # Create a virtual currency
  rc currencies create --code COINS --name "Gold Coins"

  # Interactive mode (prompts for missing fields)
  rc currencies create`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.PromptIfEmpty(&code, "Currency code", "COINS"); err != nil {
				return err
			}
			if err := cmdutil.PromptIfEmpty(&name, "Display name", "Gold Coins"); err != nil {
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
			data, err := client.Post(fmt.Sprintf("/projects/%s/virtual_currencies", url.PathEscape(pid)), map[string]any{"code": code, "name": name})
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
				t.AppendRows([]table.Row{{"Code", vc.Code}, {"Name", vc.Name}, {"State", vc.State}})
			})
			output.Success("Virtual currency created")
			output.Next("rc currencies get %s", vc.Code)
			return nil
		},
	}
	cmd.Flags().StringVar(&code, "code", "", "currency code, e.g. COINS (required)")
	cmd.Flags().StringVar(&name, "name", "", "display name (required)")
	return cmd
}

func newUpdateCmd(projectID, outputFormat *string) *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use: "update <currency-code>", Short: "Update a virtual currency",
		Example: `  # Update currency name
  rc currencies update COINS --name "Premium Coins"`,
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
			data, err := client.Post(fmt.Sprintf("/projects/%s/virtual_currencies/%s", url.PathEscape(pid), url.PathEscape(args[0])), map[string]any{"name": name})
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
				t.AppendRows([]table.Row{{"Code", vc.Code}, {"Name", vc.Name}})
			})
			output.Success("Virtual currency updated")
			output.Next("rc currencies get %s", vc.Code)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "new display name (required)")
	cmdutil.MustMarkFlagRequired(cmd, "name")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use: "delete <currency-code>", Short: "Delete a virtual currency",
		Example: `  # Delete a virtual currency
  rc currencies delete COINS`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.ConfirmDestructive("Delete", "virtual currency", args[0]); err != nil {
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
		Use: "archive <currency-code>", Short: "Archive a virtual currency",
		Example: `  # Archive a virtual currency
  rc currencies archive COINS`,
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
			_, err = client.Post(fmt.Sprintf("/projects/%s/virtual_currencies/%s/actions/archive", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Virtual currency %s archived", args[0])
			output.Next("rc currencies unarchive %s", args[0])
			return nil
		},
	}
}

func newUnarchiveCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use: "unarchive <currency-code>", Short: "Unarchive a virtual currency",
		Example: `  # Unarchive a virtual currency
  rc currencies unarchive COINS`,
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
			_, err = client.Post(fmt.Sprintf("/projects/%s/virtual_currencies/%s/actions/unarchive", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Virtual currency %s unarchived", args[0])
			output.Next("rc currencies get %s", args[0])
			return nil
		},
	}
}

func newBalanceCmd(projectID, outputFormat *string) *cobra.Command {
	var customerID string
	cmd := &cobra.Command{
		Use: "balance", Short: "Show a customer's virtual currency balances",
		Example: `  # Check balances for a customer
  rc currencies balance --customer-id user-123`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			data, err := client.Get(fmt.Sprintf("/projects/%s/customers/%s/virtual_currencies", url.PathEscape(pid), url.PathEscape(customerID)), nil)
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
	cmdutil.MustMarkFlagRequired(cmd, "customer-id")
	return cmd
}

func newCreditCmd(projectID *string) *cobra.Command {
	var (
		customerID string
		code       string
		amount     int64
		reference  string
	)
	cmd := &cobra.Command{
		Use: "credit", Short: "Create a virtual currency transaction (credit/debit)",
		Example: `  # Credit 100 coins
  rc currencies credit --customer-id user-123 --code COINS --amount 100

  # Debit 50 coins
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
				currencyAdjustmentBody(code, amount, reference),
			)
			if err != nil {
				return err
			}
			output.Success("Transaction created: %+d %s for customer %s", amount, code, customerID)
			output.Next("rc currencies balance --customer-id %s", customerID)
			return nil
		},
	}
	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringVar(&code, "code", "", "currency code (required)")
	cmd.Flags().Int64Var(&amount, "amount", 0, "amount (positive=credit, negative=debit) (required)")
	cmd.Flags().StringVar(&reference, "reference", "", "optional idempotency/reference label")
	cmdutil.MustMarkFlagRequired(cmd, "customer-id")
	cmdutil.MustMarkFlagRequired(cmd, "code")
	cmdutil.MustMarkFlagRequired(cmd, "amount")
	return cmd
}

func newUpdateBalanceCmd(projectID *string) *cobra.Command {
	var (
		customerID string
		code       string
		balance    int64
		reference  string
	)
	cmd := &cobra.Command{
		Use: "set-balance", Short: "Set a customer's virtual currency balance directly",
		Example: `  # Set balance to 500
  rc currencies set-balance --customer-id user-123 --code COINS --balance 500`,
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
				currencyAdjustmentBody(code, balance, reference),
			)
			if err != nil {
				return err
			}
			output.Success("Balance set to %d %s for customer %s", balance, code, customerID)
			output.Next("rc currencies balance --customer-id %s", customerID)
			return nil
		},
	}
	cmd.Flags().StringVar(&customerID, "customer-id", "", "customer ID (required)")
	cmd.Flags().StringVar(&code, "code", "", "currency code (required)")
	cmd.Flags().Int64Var(&balance, "balance", 0, "new balance value (required)")
	cmd.Flags().StringVar(&reference, "reference", "", "optional idempotency/reference label")
	cmdutil.MustMarkFlagRequired(cmd, "customer-id")
	cmdutil.MustMarkFlagRequired(cmd, "code")
	cmdutil.MustMarkFlagRequired(cmd, "balance")
	return cmd
}

func currencyAdjustmentBody(code string, value int64, reference string) map[string]any {
	body := map[string]any{"adjustments": map[string]int64{code: value}}
	if reference != "" {
		body["reference"] = reference
	}
	return body
}
