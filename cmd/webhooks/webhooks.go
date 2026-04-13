package webhooks

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

func NewWebhooksCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "webhooks",
		Aliases: []string{"webhook", "wh"},
		Short:   "Manage webhook integrations",
		Long: `Manage webhook integrations for a RevenueCat project.

Webhooks notify your server of subscription lifecycle events in real time.

Examples:
  rc webhooks list
  rc webhooks get whk1a2b3c4d5
  rc webhooks create --name "Production" --url https://api.example.com/webhooks/rc
  rc webhooks delete whk1a2b3c4d5`,
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newGetCmd(projectID, outputFormat))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	root.AddCommand(newUpdateCmd(projectID, outputFormat))
	root.AddCommand(newDeleteCmd(projectID))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List webhook integrations",
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			data, err := client.Get(fmt.Sprintf("/projects/%s/integrations/webhooks", url.PathEscape(pid)), nil)
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
			return nil
		},
	}
}

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <webhook-id>",
		Short: "Get a webhook by ID",
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
				t.AppendRows([]table.Row{
					{"ID", wh.ID},
					{"Name", wh.Name},
					{"URL", wh.URL},
					{"Created", output.FormatTimestamp(wh.CreatedAt)},
				})
			})
			return nil
		},
	}
}

func newCreateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		name      string
		webhookURL string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new webhook integration",
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
				fmt.Sprintf("/projects/%s/integrations/webhooks", url.PathEscape(pid)),
				map[string]any{"name": name, "url": webhookURL},
			)
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
				t.AppendRows([]table.Row{
					{"ID", wh.ID},
					{"Name", wh.Name},
					{"URL", wh.URL},
					{"Created", output.FormatTimestamp(wh.CreatedAt)},
				})
			})
			output.Success("Webhook created successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "webhook name (required)")
	cmd.Flags().StringVar(&webhookURL, "url", "", "webhook endpoint URL (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("url")
	return cmd
}

func newUpdateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		name       string
		webhookURL string
	)

	cmd := &cobra.Command{
		Use:   "update <webhook-id>",
		Short: "Update a webhook integration",
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
				t.AppendRows([]table.Row{
					{"ID", wh.ID},
					{"Name", wh.Name},
					{"URL", wh.URL},
				})
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
		Use:   "delete <webhook-id>",
		Short: "Delete a webhook integration",
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
			_, err = client.Delete(fmt.Sprintf("/projects/%s/integrations/webhooks/%s", url.PathEscape(pid), url.PathEscape(args[0])))
			if err != nil {
				return err
			}
			output.Success("Webhook %s deleted", args[0])
			return nil
		},
	}
}
