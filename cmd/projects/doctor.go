package projects

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/output"
	"github.com/andresdefi/rc/internal/projecthealth"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newDoctorCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		strict   bool
		watch    bool
		interval time.Duration
	)

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check project setup for common RevenueCat launch issues",
		Long: `Check project setup for common RevenueCat launch issues.

The doctor reads apps, products, entitlements, offerings, packages, and
package-product links. It reports missing or incomplete relationships without
mutating the project.`,
		Example: `  # Check the active project
  rc project doctor

  # Check a specific project and emit JSON
  rc project doctor --project proj1a2b3c4d5 --output json

  # Watch readiness while configuring a project
  rc project doctor --watch --interval 10s

  # Fail the command when project health has errors
  rc project doctor --strict`,
		RunE: func(c *cobra.Command, args []string) error {
			run := func(_ context.Context) error {
				pid, err := cmdutil.ResolveProject(projectID)
				if err != nil {
					return err
				}
				client, err := api.NewClient()
				if err != nil {
					return err
				}
				report, err := projecthealth.Analyze(client, pid)
				if err != nil {
					return err
				}

				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, report, renderDoctorReport(report))
				if strict && report.Status == projecthealth.StatusFail {
					return fmt.Errorf("project doctor found errors")
				}
				return nil
			}

			if watch {
				return cmdutil.Watch(c.Context(), interval, run)
			}
			return run(c.Context())
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "return a non-zero exit code when errors are found")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "continuously refresh")
	cmd.Flags().DurationVar(&interval, "interval", cmdutil.DefaultWatchInterval, "refresh interval for --watch")
	return cmd
}

func renderDoctorReport(report *projecthealth.Report) func(t table.Writer) {
	return func(t table.Writer) {
		t.AppendHeader(table.Row{"Status", "Area", "Message", "Details"})
		for _, check := range report.Checks {
			t.AppendRow(table.Row{
				strings.ToUpper(check.Status),
				check.Area,
				check.Message,
				strings.Join(check.Details, "\n"),
			})
		}
		t.AppendFooter(table.Row{
			strings.ToUpper(report.Status),
			"project",
			report.ProjectID,
			fmt.Sprintf(
				"apps=%d products=%d entitlements=%d offerings=%d packages=%d",
				report.Counts.Apps,
				report.Counts.Products,
				report.Counts.Entitlements,
				report.Counts.Offerings,
				report.Counts.Packages,
			),
		})
	}
}
