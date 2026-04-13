package projects

import (
	"encoding/json"
	"fmt"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/config"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewProjectsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "projects",
		Aliases: []string{"project", "proj"},
		Short:   "Manage RevenueCat projects",
	}

	root.AddCommand(newListCmd(outputFormat))
	root.AddCommand(newSetDefaultCmd())
	return root
}

func newListCmd(outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		RunE: func(c *cobra.Command, args []string) error {
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			data, err := client.Get("/projects", nil)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.Project]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Name", "Created"})
				for _, p := range resp.Items {
					t.AppendRow(table.Row{
						p.ID,
						p.Name,
						output.FormatTimestamp(p.CreatedAt),
					})
				}
			})
			return nil
		},
	}
}

func newSetDefaultCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-default <project-id>",
		Short: "Set the default project for all commands",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.ProjectID = args[0]
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			output.Success("Default project set to %s", args[0])
			return nil
		},
	}
}
