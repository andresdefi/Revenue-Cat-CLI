package offerings

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/completions"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewOfferingsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "offerings",
		Aliases: []string{"offering", "off"},
		Short:   "Manage offerings",
		Long: `Manage offerings in a RevenueCat project.

Offerings are the selection of products that are presented to a customer
on your paywall. Each offering contains one or more packages.

Examples:
  rc offerings list
  rc offerings get ofrnge1a2b3c4d5
  rc offerings create --lookup-key default --display-name "Standard Offering"
  rc offerings update ofrnge1a2b3c4d5 --is-current
  rc offerings archive ofrnge1a2b3c4d5`,
	}

	c := completions.OfferingIDs(projectID)
	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(completions.WithCompletion(newGetCmd(projectID, outputFormat), c))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	root.AddCommand(completions.WithCompletion(newUpdateCmd(projectID, outputFormat), c))
	root.AddCommand(completions.WithCompletion(newDeleteCmd(projectID), c))
	root.AddCommand(completions.WithCompletion(newArchiveCmd(projectID), c))
	root.AddCommand(completions.WithCompletion(newUnarchiveCmd(projectID), c))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		fetchAll bool
		limit    int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List offerings in a project",
		Example: `  # List all offerings
  rc offerings list

  # List offerings for a specific project as JSON
  rc offerings list --project proj1a2b3c4d5 --output json

  # Use a production profile
  rc offerings list --profile production

  # Find the current offering
  rc offerings list --output json | jq -r '.items[] | select(.is_current) | .id'

  # List offerings, then inspect packages on one offering
  rc offerings list
  rc offerings get ofrnge1a2b3c4d5 --output json | jq '.packages.items'

  # Fetch every page
  rc offerings list --all --limit 100`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			path := fmt.Sprintf("/projects/%s/offerings", url.PathEscape(pid))
			query := url.Values{}
			query.Set("expand", "items.package")
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}

			if fetchAll {
				items, err := api.PaginateAll[api.Offering](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Lookup Key", "Display Name", "Current", "State", "Packages", "Created"})
					for _, o := range items {
						pkgCount := 0
						if o.Packages != nil {
							pkgCount = len(o.Packages.Items)
						}
						current := ""
						if o.IsCurrent {
							current = "yes"
						}
						t.AppendRow(table.Row{o.ID, o.LookupKey, o.DisplayName, current, o.State, pkgCount, output.FormatTimestamp(o.CreatedAt)})
					}
					t.AppendFooter(table.Row{"", "", "", "", "", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}

			data, err := client.Get(path, query)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.Offering]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Lookup Key", "Display Name", "Current", "State", "Packages", "Created"})
				for _, o := range resp.Items {
					pkgCount := 0
					if o.Packages != nil {
						pkgCount = len(o.Packages.Items)
					}
					current := ""
					if o.IsCurrent {
						current = "yes"
					}
					t.AppendRow(table.Row{o.ID, o.LookupKey, o.DisplayName, current, o.State, pkgCount, output.FormatTimestamp(o.CreatedAt)})
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
	return cmd
}

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <offering-id>",
		Short: "Get an offering by ID",
		Example: `  # Get offering details with packages
  rc offerings get ofrnge1a2b3c4d5

  # Get as JSON
  rc offerings get ofrnge1a2b3c4d5 -o json`,
		Args: cobra.ExactArgs(1),
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
			query.Set("expand", "package,package.product")

			data, err := client.Get(fmt.Sprintf("/projects/%s/offerings/%s", url.PathEscape(pid), url.PathEscape(args[0])), query)
			if err != nil {
				return err
			}

			var offering api.Offering
			if err := json.Unmarshal(data, &offering); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, offering, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", offering.ID},
					{"Lookup Key", offering.LookupKey},
					{"Display Name", offering.DisplayName},
					{"Current", offering.IsCurrent},
					{"State", offering.State},
					{"Created", output.FormatTimestamp(offering.CreatedAt)},
				})
				if offering.Packages != nil && len(offering.Packages.Items) > 0 {
					t.AppendSeparator()
					t.AppendRow(table.Row{"PACKAGES", ""})
					for _, p := range offering.Packages.Items {
						pos := "-"
						if p.Position != nil {
							pos = fmt.Sprintf("%d", *p.Position)
						}
						t.AppendRow(table.Row{fmt.Sprintf("  %s (%s)", p.LookupKey, p.ID), fmt.Sprintf("pos: %s", pos)})
					}
				}
			})
			return nil
		},
	}
}

func newCreateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		lookupKey   string
		displayName string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new offering",
		Long: `Create a new offering. Required flags are prompted interactively when
running in a terminal and not provided on the command line.`,
		Example: `  # Create an offering
  rc offerings create --lookup-key default --display-name "Standard Offering"

  # Create and print JSON
  rc offerings create --lookup-key winback --display-name "Winback Offer" --output json

  # Use a staging profile
  rc offerings create --lookup-key beta --display-name "Beta Offer" --profile staging

  # Capture the offering ID
  rc offerings create --lookup-key default --display-name "Standard Offering" --output json | jq -r '.id'

  # Create an offering, then add a package
  rc offerings create --lookup-key default --display-name "Standard Offering"
  rc packages create --offering-id ofrnge1a2b3c --lookup-key monthly --display-name "Monthly"

  # Interactive mode (prompts for missing fields)
  rc offerings create`,
		RunE: func(c *cobra.Command, args []string) error {
			// Interactive prompts for missing required fields
			if err := cmdutil.PromptIfEmpty(&lookupKey, "Lookup key", "default"); err != nil {
				return err
			}
			if err := cmdutil.PromptIfEmpty(&displayName, "Display name", "Standard Offering"); err != nil {
				return err
			}

			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			body := map[string]any{"lookup_key": lookupKey, "display_name": displayName}
			data, err := client.Post(fmt.Sprintf("/projects/%s/offerings", url.PathEscape(pid)), body)
			if err != nil {
				return err
			}

			var offering api.Offering
			if err := json.Unmarshal(data, &offering); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, offering, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", offering.ID},
					{"Lookup Key", offering.LookupKey},
					{"Display Name", offering.DisplayName},
					{"Current", offering.IsCurrent},
					{"State", offering.State},
					{"Created", output.FormatTimestamp(offering.CreatedAt)},
				})
			})
			output.Success("Offering created successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&lookupKey, "lookup-key", "", "lookup key identifier")
	cmd.Flags().StringVar(&displayName, "display-name", "", "display name")
	return cmd
}

func newUpdateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		displayName string
		isCurrent   bool
	)

	cmd := &cobra.Command{
		Use:   "update <offering-id>",
		Short: "Update an offering",
		Example: `  # Update display name
  rc offerings update ofrnge1a2b3c4d5 --display-name "Premium Offering"

  # Set as current offering
  rc offerings update ofrnge1a2b3c4d5 --is-current`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			body := map[string]any{}
			if c.Flags().Changed("display-name") {
				body["display_name"] = displayName
			}
			if c.Flags().Changed("is-current") {
				body["is_current"] = isCurrent
			}

			data, err := client.Post(fmt.Sprintf("/projects/%s/offerings/%s", url.PathEscape(pid), url.PathEscape(args[0])), body)
			if err != nil {
				return err
			}

			var offering api.Offering
			if err := json.Unmarshal(data, &offering); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, offering, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", offering.ID},
					{"Display Name", offering.DisplayName},
					{"Current", offering.IsCurrent},
					{"State", offering.State},
				})
			})
			output.Success("Offering updated")
			return nil
		},
	}

	cmd.Flags().StringVar(&displayName, "display-name", "", "new display name")
	cmd.Flags().BoolVar(&isCurrent, "is-current", false, "set as current offering")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <offering-id>",
		Short: "Delete an offering and its packages",
		Example: `  # Delete an offering
  rc offerings delete ofrnge1a2b3c4d5`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.ConfirmDestructive("Delete", "offering", args[0]); err != nil {
				return err
			}
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			_, err = client.Delete(fmt.Sprintf("/projects/%s/offerings/%s", url.PathEscape(pid), url.PathEscape(args[0])))
			if err != nil {
				return err
			}
			output.Success("Offering %s deleted", args[0])
			return nil
		},
	}
}

func newArchiveCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <offering-id>",
		Short: "Archive an offering",
		Example: `  # Archive an offering
  rc offerings archive ofrnge1a2b3c4d5`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			_, err = client.Post(fmt.Sprintf("/projects/%s/offerings/%s/actions/archive", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Offering %s archived", args[0])
			return nil
		},
	}
}

func newUnarchiveCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "unarchive <offering-id>",
		Short: "Unarchive an offering",
		Example: `  # Unarchive an offering
  rc offerings unarchive ofrnge1a2b3c4d5`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}
			_, err = client.Post(fmt.Sprintf("/projects/%s/offerings/%s/actions/unarchive", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}
			output.Success("Offering %s unarchived", args[0])
			return nil
		},
	}
}
