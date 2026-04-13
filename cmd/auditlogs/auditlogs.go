package auditlogs

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

func NewAuditLogsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "audit-logs",
		Aliases: []string{"audit", "logs"},
		Short:   "View audit logs",
		Long: `View audit logs for a RevenueCat project.

Audit logs track changes made to your project configuration.

Examples:
  rc audit-logs list
  rc audit-logs list --start-date 2024-01-01 --end-date 2024-01-31`,
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		startDate string
		endDate   string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List audit log entries",
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
			if startDate != "" {
				query.Set("start_date", startDate)
			}
			if endDate != "" {
				query.Set("end_date", endDate)
			}

			data, err := client.Get(fmt.Sprintf("/projects/%s/audit_logs", url.PathEscape(pid)), query)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.AuditLogEntry]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Action", "Actor", "Details", "Created"})
				for _, e := range resp.Items {
					details := e.Details
					if len(details) > 60 {
						details = details[:57] + "..."
					}
					t.AppendRow(table.Row{e.ID, e.Action, e.Actor, details, output.FormatTimestamp(e.CreatedAt)})
				}
			})
			return nil
		},
	}

	cmd.Flags().StringVar(&startDate, "start-date", "", "filter from date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "filter to date (YYYY-MM-DD)")
	return cmd
}
