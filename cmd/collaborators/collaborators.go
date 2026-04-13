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
		Long: `View collaborators who have access to a RevenueCat project.

Examples:
  rc collaborators list`,
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List project collaborators",
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			data, err := client.Get(fmt.Sprintf("/projects/%s/collaborators", url.PathEscape(pid)), nil)
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
				for _, c := range resp.Items {
					t.AppendRow(table.Row{c.ID, c.Email, c.Role})
				}
			})
			return nil
		},
	}
}
