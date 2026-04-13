package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/andresdefi/rc/cmd/apps"
	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/cmd/auditlogs"
	"github.com/andresdefi/rc/cmd/auth"
	"github.com/andresdefi/rc/cmd/charts"
	"github.com/andresdefi/rc/cmd/collaborators"
	"github.com/andresdefi/rc/cmd/currencies"
	"github.com/andresdefi/rc/cmd/customers"
	"github.com/andresdefi/rc/cmd/entitlements"
	"github.com/andresdefi/rc/cmd/offerings"
	"github.com/andresdefi/rc/cmd/packages"
	"github.com/andresdefi/rc/cmd/paywalls"
	"github.com/andresdefi/rc/cmd/products"
	"github.com/andresdefi/rc/cmd/projects"
	"github.com/andresdefi/rc/cmd/purchases"
	"github.com/andresdefi/rc/cmd/subscriptions"
	"github.com/andresdefi/rc/cmd/webhooks"
	"github.com/spf13/cobra"
)

var (
	projectID    string
	outputFormat string
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
		SilenceUsage:          true,
		SilenceErrors:         true,
		SuggestionsMinimumDistance: 2,
	}

	root.PersistentFlags().StringVarP(&projectID, "project", "p", "", "project ID (overrides default project)")
	root.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "output format: table, json (default: table for TTY, json for pipes)")

	// Meta
	root.AddCommand(newVersionCmd())

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

	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)

		var apiErr *api.Error
		switch {
		case errors.As(err, &apiErr) && apiErr.Type == "authentication_error":
			os.Exit(3) // auth error
		case errors.As(err, &apiErr):
			os.Exit(4) // API error
		default:
			os.Exit(1) // general error
		}
	}
}
