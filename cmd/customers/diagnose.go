package customers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/customerdiagnosis"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newDiagnoseCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		strict   bool
		watch    bool
		interval time.Duration
	)

	cmd := &cobra.Command{
		Use:   "diagnose <customer-id>",
		Short: "Diagnose why a customer does or does not have access",
		Long: `Diagnose why a customer does or does not have access.

The diagnosis is read-only. It looks up the customer, active entitlements,
subscriptions, purchases, aliases, and attributes, then reports likely access
issues and follow-up commands for support debugging.`,
		Example: `  # Diagnose customer access
  rc customers diagnose user-123

  # Emit JSON for scripts
  rc customers diagnose user-123 --output json

  # Watch while a customer restores purchases or changes state
  rc customers diagnose user-123 --watch --interval 10s

  # Fail when blocking access findings are found
  rc customers diagnose user-123 --strict`,
		Args: cobra.ExactArgs(1),
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
				report, err := customerdiagnosis.Analyze(client, pid, args[0])
				if err != nil {
					return err
				}

				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, report, renderDiagnosisReport(report))
				if strict && report.Status == customerdiagnosis.StatusFail {
					return fmt.Errorf("customer diagnosis found failed checks")
				}
				return nil
			}

			if watch {
				return cmdutil.Watch(c.Context(), interval, run)
			}
			return run(c.Context())
		},
	}

	cmd.Flags().BoolVar(&strict, "strict", false, "return a non-zero exit code when failed checks are found")
	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "continuously refresh")
	cmd.Flags().DurationVar(&interval, "interval", cmdutil.DefaultWatchInterval, "refresh interval for --watch")
	return cmd
}

func renderDiagnosisReport(report *customerdiagnosis.Report) func(t table.Writer) {
	return func(t table.Writer) {
		t.AppendHeader(table.Row{"Section", "Status", "Item", "Details"})
		t.AppendRow(table.Row{
			"summary",
			strings.ToUpper(report.AccessSummary),
			report.CustomerID,
			fmt.Sprintf(
				"entitlements=%d subscriptions=%d purchases=%d aliases=%d",
				report.Counts.ActiveEntitlements,
				report.Counts.Subscriptions,
				report.Counts.Purchases,
				report.Counts.Aliases,
			),
		})

		appendEntitlementRows(t, report.ActiveEntitlements)
		appendSubscriptionRows(t, report.Subscriptions)
		appendPurchaseRows(t, report.Purchases)
		appendAliasRows(t, report.Aliases)

		t.AppendSeparator()
		for _, finding := range report.Findings {
			t.AppendRow(table.Row{
				"finding",
				strings.ToUpper(finding.Severity),
				finding.Area,
				finding.Message + detailSuffix(finding.Details),
			})
		}

		if len(report.NextCommands) > 0 {
			t.AppendSeparator()
			for _, command := range report.NextCommands {
				t.AppendRow(table.Row{"next", "", command, ""})
			}
		}

		t.AppendFooter(table.Row{
			"status",
			strings.ToUpper(report.Status),
			report.AccessSummary,
			fmt.Sprintf("%d finding(s)", len(report.Findings)),
		})
	}
}

func appendEntitlementRows(t table.Writer, entitlements []customerdiagnosis.EntitlementSummary) {
	if len(entitlements) == 0 {
		t.AppendRow(table.Row{"entitlement", "FAIL", "(none)", "no active entitlements"})
		return
	}
	for _, entitlement := range entitlements {
		t.AppendRow(table.Row{
			"entitlement",
			"PASS",
			entitlement.EntitlementID,
			"expires: " + formatOptionalTimestamp(entitlement.ExpiresAt),
		})
	}
}

func appendSubscriptionRows(t table.Writer, subscriptions []customerdiagnosis.SubscriptionSummary) {
	if len(subscriptions) == 0 {
		t.AppendRow(table.Row{"subscription", "INFO", "(none)", ""})
		return
	}
	for _, subscription := range subscriptions {
		t.AppendRow(table.Row{
			"subscription",
			strings.ToUpper(subscription.Status),
			subscription.ID,
			fmt.Sprintf(
				"product=%s store=%s gives_access=%t renewal=%s period_ends=%s",
				subscription.ProductID,
				subscription.Store,
				subscription.GivesAccess,
				subscription.AutoRenewalStatus,
				formatOptionalTimestamp(subscription.CurrentPeriodEndsAt),
			),
		})
	}
}

func appendPurchaseRows(t table.Writer, purchases []customerdiagnosis.PurchaseSummary) {
	if len(purchases) == 0 {
		t.AppendRow(table.Row{"purchase", "INFO", "(none)", ""})
		return
	}
	for _, purchase := range purchases {
		t.AppendRow(table.Row{
			"purchase",
			strings.ToUpper(purchase.Status),
			purchase.ID,
			fmt.Sprintf(
				"product=%s store=%s quantity=%d purchased=%s",
				purchase.ProductID,
				purchase.Store,
				purchase.Quantity,
				output.FormatTimestamp(purchase.PurchasedAt),
			),
		})
	}
}

func appendAliasRows(t table.Writer, aliases []string) {
	if len(aliases) == 0 {
		t.AppendRow(table.Row{"alias", "INFO", "(none)", ""})
		return
	}
	for _, alias := range aliases {
		t.AppendRow(table.Row{"alias", "INFO", alias, "may explain split identity"})
	}
}

func detailSuffix(details []string) string {
	if len(details) == 0 {
		return ""
	}
	return "\n" + strings.Join(details, "\n")
}
