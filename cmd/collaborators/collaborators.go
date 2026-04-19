package collaborators

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

func NewCollaboratorsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "collaborators",
		Aliases: []string{"collaborator", "collab"},
		Short:   "View project collaborators",
	}
	root.AddCommand(newListCmd(projectID, outputFormat))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		fetchAll bool
		limit    int
	)
	cmd := &cobra.Command{
		Use: "list", Short: "List project collaborators",
		Example: `  # List collaborators
  rc collaborators list

  # List with JSON output
  rc collaborators list -o json`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/projects/%s/collaborators", url.PathEscape(pid))
			query := url.Values{}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}
			if fetchAll {
				items, err := api.PaginateAll[api.Collaborator](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Email", "Role"})
					for _, co := range items {
						t.AppendRow(table.Row{co.ID, co.Email, co.Role})
					}
					t.AppendFooter(table.Row{"", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}
			data, err := client.Get(path, query)
			if err != nil {
				return err
			}
			var resp api.ListResponse[api.Collaborator]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}
			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Email", "Role"})
				for _, co := range resp.Items {
					t.AppendRow(table.Row{co.ID, co.Email, co.Role})
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
	cmdutil.SetFieldsPreset(cmd, []string{"id", "email", "role"})
	return cmd
}
