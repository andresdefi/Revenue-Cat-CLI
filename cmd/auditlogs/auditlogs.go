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
	}
	root.AddCommand(newListCmd(projectID, outputFormat))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		startDate string
		endDate   string
		fetchAll  bool
		limit     int
	)
	cmd := &cobra.Command{
		Use: "list", Short: "List audit log entries",
		Example: `  # List recent audit logs
  rc audit-logs list

  # Filter by date range
  rc audit-logs list --start-date 2024-01-01 --end-date 2024-01-31

  # Fetch all pages
  rc audit-logs list --all`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/projects/%s/audit_logs", url.PathEscape(pid))
			query := url.Values{}
			if startDate != "" {
				query.Set("start_date", startDate)
			}
			if endDate != "" {
				query.Set("end_date", endDate)
			}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}
			if fetchAll {
				items, err := api.PaginateAll[api.AuditLogEntry](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Action", "Actor", "Details", "Created"})
					for _, e := range items {
						details := e.Details
						if len(details) > 60 {
							details = details[:57] + "..."
						}
						t.AppendRow(table.Row{e.ID, e.Action, e.Actor, details, output.FormatTimestamp(e.CreatedAt)})
					}
					t.AppendFooter(table.Row{"", "", "", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}
			data, err := client.Get(path, query)
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
			if resp.NextPage != nil {
				output.Warn("More results available (use --all for more)")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&startDate, "start-date", "", "filter from date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "filter to date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&fetchAll, "all", false, "fetch all pages")
	cmd.Flags().IntVar(&limit, "limit", 0, "max items per page")
	cmdutil.SetFieldsPreset(cmd, []string{"id", "created_at", "actor", "action"})
	return cmd
}
