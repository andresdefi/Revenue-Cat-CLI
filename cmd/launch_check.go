package cmd

import (
	"fmt"
	"strings"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/output"
	"github.com/andresdefi/rc/internal/projecthealth"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newLaunchCheckCmd(projectID, outputFormat *string) *cobra.Command {
	var strict bool

	cmd := &cobra.Command{
		Use:   "launch-check",
		Short: "Run a pre-launch RevenueCat readiness check",
		Long: `Run a pre-launch RevenueCat readiness check.

Launch check reuses the project health analyzer and summarizes whether the
active project has the minimum product, entitlement, offering, package, and
package-product paths needed before shipping.`,
		Example: `  # Check whether the active project is ready to launch
  rc launch-check

  # Emit JSON for automation
  rc launch-check --output json

  # Fail when required launch paths are missing
  rc launch-check --strict`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			health, err := projecthealth.Analyze(client, pid)
			if err != nil {
				return err
			}
			report := projecthealth.AssessLaunch(health)

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, report, renderLaunchReport(report))
			if strict && !report.Ready {
				return fmt.Errorf("launch check failed")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "return a non-zero exit code when launch requirements are missing")
	return cmd
}

func renderLaunchReport(report *projecthealth.LaunchReport) func(t table.Writer) {
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
		ready := "yes"
		if !report.Ready {
			ready = "no"
		}
		t.AppendFooter(table.Row{
			strings.ToUpper(report.Status),
			"launch",
			fmt.Sprintf("ready: %s", ready),
			fmt.Sprintf(
				"apps=%d active_products=%d active_entitlements=%d current_offerings=%d packages=%d package_products=%d",
				report.Counts.Apps,
				report.Counts.ActiveProducts,
				report.Counts.ActiveEntitlements,
				report.Counts.CurrentOfferings,
				report.Counts.Packages,
				report.Counts.PackageProductLinks,
			),
		})
	}
}
