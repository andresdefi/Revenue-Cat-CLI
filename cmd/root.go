package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/andresdefi/rc/cmd/apps"
	"github.com/andresdefi/rc/cmd/auditlogs"
	"github.com/andresdefi/rc/cmd/auth"
	"github.com/andresdefi/rc/cmd/charts"
	"github.com/andresdefi/rc/cmd/collaborators"
	"github.com/andresdefi/rc/cmd/currencies"
	"github.com/andresdefi/rc/cmd/customers"
	"github.com/andresdefi/rc/cmd/entitlements"
	mcpcmd "github.com/andresdefi/rc/cmd/mcp"
	"github.com/andresdefi/rc/cmd/offerings"
	"github.com/andresdefi/rc/cmd/packages"
	"github.com/andresdefi/rc/cmd/paywalls"
	"github.com/andresdefi/rc/cmd/products"
	"github.com/andresdefi/rc/cmd/projects"
	"github.com/andresdefi/rc/cmd/purchases"
	"github.com/andresdefi/rc/cmd/subscriptions"
	"github.com/andresdefi/rc/cmd/transfer"
	"github.com/andresdefi/rc/cmd/webhooks"
	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/exitcode"
	"github.com/andresdefi/rc/internal/output"
	"github.com/andresdefi/rc/internal/update"
	"github.com/andresdefi/rc/internal/version"
	"github.com/spf13/cobra"
)

var (
	projectID    string
	outputFormat string
	profileFlag  string
	noColorFlag  bool
	prettyFlag   bool
	verboseFlag  bool
	quietFlag    bool
	dryRunFlag   bool
	forceFlag    bool
	fieldsFlag   string
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "rc",
		Short: "RevenueCat CLI - manage your RevenueCat projects from the terminal",
		Long: `rc is an unofficial CLI for the RevenueCat REST API v2.

Manage products, entitlements, offerings, customers, subscriptions,
and more without leaving your terminal.

Get started:
  rc auth login           Authenticate with your API key
  rc projects list        List your projects
  rc products list        List products in a project
  rc customers lookup     Look up a customer
  rc charts overview      View metrics overview

Full API v2 coverage: projects, apps, products, entitlements, offerings,
packages, customers, subscriptions, purchases, webhooks, charts, paywalls,
audit logs, collaborators, and virtual currencies.`,
		Version:                    version.Version,
		SilenceUsage:               true,
		SilenceErrors:              true,
		SuggestionsMinimumDistance: 2,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if outputFormat != "" && outputFormat != "table" && outputFormat != "json" {
				return fmt.Errorf("invalid output format %q (must be table or json)", outputFormat)
			}
			if noColorFlag {
				output.ColorDisabled = true
			}
			if prettyFlag {
				output.PrettyJSON = true
			}
			if verboseFlag {
				output.Verbose = true
			}
			if quietFlag {
				output.Quiet = true
			}
			if dryRunFlag {
				api.DryRun = true
			}
			cmdutil.ActiveProfile = profileFlag
			if forceFlag {
				cmdutil.ForceYes = true
			}
			cmdutil.FieldsFlag = fieldsFlag
			output.FieldsFilter = fieldsFlag
			return nil
		},
	}

	root.SetVersionTemplate("rc {{.Version}}\n")

	root.PersistentFlags().StringVarP(&projectID, "project", "p", "", "project ID (overrides default project)")
	root.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "output format: table, json (default: table for TTY, json for pipes)")
	root.PersistentFlags().StringVar(&profileFlag, "profile", "", "config profile to use (overrides RC_PROFILE and current_profile)")
	root.PersistentFlags().BoolVar(&noColorFlag, "no-color", false, "disable color output (also respects NO_COLOR env var)")
	root.PersistentFlags().BoolVar(&prettyFlag, "pretty", false, "pretty-print JSON output (default for TTY, compact for pipes)")
	root.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "show HTTP request/response details for debugging")
	root.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress non-essential output (success messages, warnings, progress)")
	root.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "show what would be done without executing mutations")
	root.PersistentFlags().BoolVarP(&forceFlag, "yes", "y", false, "skip confirmation prompts for destructive operations")
	root.PersistentFlags().StringVar(&fieldsFlag, "fields", "", "comma-separated list of fields to include in JSON output")

	// Meta
	root.AddCommand(newVersionCmd())
	root.AddCommand(newCompletionCmd())

	// Auth (no project needed)
	root.AddCommand(auth.NewAuthCmd())

	// Project management
	root.AddCommand(projects.NewProjectsCmd(&projectID, &outputFormat))
	root.AddCommand(apps.NewAppsCmd(&projectID, &outputFormat))
	root.AddCommand(collaborators.NewCollaboratorsCmd(&projectID, &outputFormat))

	// Product configuration
	root.AddCommand(products.NewProductsCmd(&projectID, &outputFormat))
	root.AddCommand(entitlements.NewEntitlementsCmd(&projectID, &outputFormat))
	root.AddCommand(offerings.NewOfferingsCmd(&projectID, &outputFormat))
	root.AddCommand(packages.NewPackagesCmd(&projectID, &outputFormat))

	// Customer data
	root.AddCommand(customers.NewCustomersCmd(&projectID, &outputFormat))
	root.AddCommand(subscriptions.NewSubscriptionsCmd(&projectID, &outputFormat))
	root.AddCommand(purchases.NewPurchasesCmd(&projectID, &outputFormat))

	// Integrations & analytics
	root.AddCommand(webhooks.NewWebhooksCmd(&projectID, &outputFormat))
	root.AddCommand(charts.NewChartsCmd(&projectID, &outputFormat))
	root.AddCommand(paywalls.NewPaywallsCmd(&projectID, &outputFormat))
	root.AddCommand(auditlogs.NewAuditLogsCmd(&projectID, &outputFormat))
	root.AddCommand(currencies.NewCurrenciesCmd(&projectID, &outputFormat))

	// MCP server
	root.AddCommand(mcpcmd.NewMCPCmd())

	// Project config transfer
	root.AddCommand(transfer.NewExportCmd(&projectID, &outputFormat))
	root.AddCommand(transfer.NewImportCmd(&projectID, &outputFormat))

	return root
}

func Execute() {
	updateCh := make(chan string, 1)
	update.CheckAsync(updateCh)
	defer update.PrintNotice(updateCh)

	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%sError:%s %s\n", output.ColorRed(), output.ColorReset(), err)

		var apiErr *api.Error
		switch {
		case errors.As(err, &apiErr) && apiErr.StatusCode == 401:
			os.Exit(exitcode.AuthError)
		case errors.As(err, &apiErr) && apiErr.StatusCode == 404:
			os.Exit(exitcode.NotFoundError)
		case errors.As(err, &apiErr) && apiErr.StatusCode == 429:
			os.Exit(exitcode.RateLimitError)
		case errors.As(err, &apiErr) && apiErr.StatusCode >= 500:
			os.Exit(exitcode.ServerError)
		case errors.As(err, &apiErr):
			os.Exit(exitcode.APIError)
		default:
			os.Exit(exitcode.GeneralError)
		}
	}
}
