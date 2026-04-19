package paywalls

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

func NewPaywallsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "paywalls",
		Aliases: []string{"paywall"},
		Short:   "Manage paywalls",
	}
	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newGetCmd(projectID, outputFormat))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	root.AddCommand(newValidateCmd(projectID, outputFormat))
	root.AddCommand(newDeleteCmd(projectID))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		fetchAll bool
		limit    int
	)
	cmd := &cobra.Command{
		Use: "list", Short: "List paywalls",
		Example: `  # List paywalls
  rc paywalls list

  # List with JSON output
  rc paywalls list -o json`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/projects/%s/paywalls", url.PathEscape(pid))
			query := url.Values{}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}
			if fetchAll {
				items, err := api.PaginateAll[api.Paywall](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Created"})
					for _, p := range items {
						t.AppendRow(table.Row{p.ID, output.FormatTimestamp(p.CreatedAt)})
					}
					t.AppendFooter(table.Row{"", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}
			data, err := client.Get(path, query)
			if err != nil {
				return err
			}
			var resp api.ListResponse[api.Paywall]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Created"})
				for _, p := range resp.Items {
					t.AppendRow(table.Row{p.ID, output.FormatTimestamp(p.CreatedAt)})
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
		Use: "get <paywall-id>", Short: "Get a paywall by ID",
		Example: `  # Get paywall details
  rc paywalls get pw1a2b3c4d5

  # Get as JSON
  rc paywalls get pw1a2b3c4d5 -o json`,
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
			data, err := client.Get(fmt.Sprintf("/projects/%s/paywalls/%s", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			var pw api.Paywall
			if err := json.Unmarshal(data, &pw); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, pw, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{{"ID", pw.ID}, {"Created", output.FormatTimestamp(pw.CreatedAt)}})
			})
			return nil
		},
	}
}

func newCreateCmd(projectID, outputFormat *string) *cobra.Command {
	var offeringID string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a paywall",
		Long: `Create a paywall. Required flags are prompted interactively when
running in a terminal and not provided on the command line.`,
		Example: `  # Create a paywall
  rc paywalls create --offering-id ofrnge1a2b3c4d5

  # Interactive mode (prompts for missing fields)
  rc paywalls create`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.PromptIfEmpty(&offeringID, "Offering ID", "ofrnge1a2b3c4d5"); err != nil {
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
			data, err := client.Post(fmt.Sprintf("/projects/%s/paywalls", url.PathEscape(pid)), map[string]any{"offering_id": offeringID})
			if err != nil {
				return err
			}
			var pw api.Paywall
			if err := json.Unmarshal(data, &pw); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, pw, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{{"ID", pw.ID}, {"Created", output.FormatTimestamp(pw.CreatedAt)}})
			})
			output.Success("Paywall created")
			output.Next("rc paywalls get %s", pw.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&offeringID, "offering-id", "", "offering ID for the paywall (required)")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use: "delete <paywall-id>", Short: "Delete a paywall",
		Example: `  # Delete a paywall
  rc paywalls delete pw1a2b3c4d5`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.ConfirmDestructive("Delete", "paywall", args[0]); err != nil {
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
			_, err = client.Delete(fmt.Sprintf("/projects/%s/paywalls/%s", url.PathEscape(pid), url.PathEscape(args[0])))
			if err != nil {
				return err
			}
			output.Success("Paywall %s deleted", args[0])
			return nil
		},
	}
}
