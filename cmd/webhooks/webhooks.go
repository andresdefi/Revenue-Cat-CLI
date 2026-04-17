package webhooks

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/completions"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewWebhooksCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "webhooks",
		Aliases: []string{"webhook", "wh"},
		Short:   "Manage webhook integrations",
	}
	c := completions.WebhookIDs(projectID)
	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(completions.WithCompletion(newGetCmd(projectID, outputFormat), c))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	root.AddCommand(completions.WithCompletion(newUpdateCmd(projectID, outputFormat), c))
	root.AddCommand(completions.WithCompletion(newDeleteCmd(projectID), c))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		fetchAll bool
		limit    int
	)
	cmd := &cobra.Command{
		Use: "list", Short: "List webhook integrations",
		Example: `  # List webhooks
  rc webhooks list

  # List with JSON output
  rc webhooks list -o json`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/projects/%s/integrations/webhooks", url.PathEscape(pid))
			query := url.Values{}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}
			if fetchAll {
				items, err := api.PaginateAll[api.Webhook](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Name", "URL", "Created"})
					for _, w := range items {
						t.AppendRow(table.Row{w.ID, w.Name, w.URL, output.FormatTimestamp(w.CreatedAt)})
					}
					t.AppendFooter(table.Row{"", "", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}
			data, err := client.Get(path, query)
			if err != nil {
				return err
			}
			var resp api.ListResponse[api.Webhook]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Name", "URL", "Created"})
				for _, w := range resp.Items {
					t.AppendRow(table.Row{w.ID, w.Name, w.URL, output.FormatTimestamp(w.CreatedAt)})
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
		Use: "get <webhook-id>", Short: "Get a webhook by ID",
		Example: `  # Get webhook details
  rc webhooks get wh1a2b3c4d5`,
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
			data, err := client.Get(fmt.Sprintf("/projects/%s/integrations/webhooks/%s", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			var wh api.Webhook
			if err := json.Unmarshal(data, &wh); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, wh, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{{"ID", wh.ID}, {"Name", wh.Name}, {"URL", wh.URL}, {"Created", output.FormatTimestamp(wh.CreatedAt)}})
			})
			return nil
		},
	}
}

func newCreateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		name       string
		webhookURL string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new webhook integration",
		Long: `Create a new webhook integration. Required flags are prompted
interactively when running in a terminal and not provided on the command line.`,
		Example: `  # Create a webhook
  rc webhooks create --name "My Webhook" --url https://example.com/webhook

  # Interactive mode (prompts for missing fields)
  rc webhooks create`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.PromptIfEmpty(&name, "Webhook name", "My Webhook"); err != nil {
				return err
			}
			if err := cmdutil.PromptIfEmpty(&webhookURL, "Webhook URL", "https://example.com/webhook"); err != nil {
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
			data, err := client.Post(fmt.Sprintf("/projects/%s/integrations/webhooks", url.PathEscape(pid)), map[string]any{"name": name, "url": webhookURL})
			if err != nil {
				return err
			}
			var wh api.Webhook
			if err := json.Unmarshal(data, &wh); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, wh, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{{"ID", wh.ID}, {"Name", wh.Name}, {"URL", wh.URL}, {"Created", output.FormatTimestamp(wh.CreatedAt)}})
			})
			output.Success("Webhook created successfully")
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "webhook name (required)")
	cmd.Flags().StringVar(&webhookURL, "url", "", "webhook endpoint URL (required)")
	return cmd
}

func newUpdateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		name       string
		webhookURL string
	)
	cmd := &cobra.Command{
		Use: "update <webhook-id>", Short: "Update a webhook integration",
		Example: `  # Update webhook name
  rc webhooks update wh1a2b3c4d5 --name "Updated Webhook"

  # Update webhook URL
  rc webhooks update wh1a2b3c4d5 --url https://example.com/new-webhook`,
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
			body := map[string]any{}
			if c.Flags().Changed("name") {
				body["name"] = name
			}
			if c.Flags().Changed("url") {
				body["url"] = webhookURL
			}
			data, err := client.Post(fmt.Sprintf("/projects/%s/integrations/webhooks/%s", url.PathEscape(pid), url.PathEscape(args[0])), body)
			if err != nil {
				return err
			}
			var wh api.Webhook
			if err := json.Unmarshal(data, &wh); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, wh, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{{"ID", wh.ID}, {"Name", wh.Name}, {"URL", wh.URL}})
			})
			output.Success("Webhook updated")
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "new webhook name")
	cmd.Flags().StringVar(&webhookURL, "url", "", "new webhook URL")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use: "delete <webhook-id>", Short: "Delete a webhook integration",
		Example: `  # Delete a webhook
  rc webhooks delete wh1a2b3c4d5`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.ConfirmDestructive("Delete", "webhook", args[0]); err != nil {
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
			_, err = client.Delete(fmt.Sprintf("/projects/%s/integrations/webhooks/%s", url.PathEscape(pid), url.PathEscape(args[0])))
			if err != nil {
				return err
			}
			output.Success("Webhook %s deleted", args[0])
			return nil
		},
	}
}
