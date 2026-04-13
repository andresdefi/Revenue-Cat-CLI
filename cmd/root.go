package cmd

import (
	"fmt"
	"os"

	"github.com/andresdefi/rc/cmd/auth"
	"github.com/andresdefi/rc/cmd/customers"
	"github.com/andresdefi/rc/cmd/entitlements"
	"github.com/andresdefi/rc/cmd/offerings"
	"github.com/andresdefi/rc/cmd/products"
	"github.com/andresdefi/rc/cmd/projects"
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

Manage products, entitlements, offerings, and customers without
leaving your terminal. Built for developers who prefer the command line.

Get started:
  rc auth login        Authenticate with your API key
  rc projects list     List your projects
  rc products list     List products in a project`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().StringVarP(&projectID, "project", "p", "", "project ID (overrides default project)")
	root.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format: table, json")

	root.AddCommand(auth.NewAuthCmd())
	root.AddCommand(projects.NewProjectsCmd(&projectID, &outputFormat))
	root.AddCommand(products.NewProductsCmd(&projectID, &outputFormat))
	root.AddCommand(entitlements.NewEntitlementsCmd(&projectID, &outputFormat))
	root.AddCommand(offerings.NewOfferingsCmd(&projectID, &outputFormat))
	root.AddCommand(customers.NewCustomersCmd(&projectID, &outputFormat))

	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
