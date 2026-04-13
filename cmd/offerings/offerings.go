package offerings

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

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newGetCmd(projectID, outputFormat))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	root.AddCommand(newUpdateCmd(projectID, outputFormat))
	root.AddCommand(newDeleteCmd(projectID))
	root.AddCommand(newArchiveCmd(projectID))
	root.AddCommand(newUnarchiveCmd(projectID))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List offerings in a project",
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
			query.Set("expand", "items.package")

			data, err := client.Get(fmt.Sprintf("/projects/%s/offerings", url.PathEscape(pid)), query)
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
			return nil
		},
	}
}

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <offering-id>",
		Short: "Get an offering by ID",
		Args:  cobra.ExactArgs(1),
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
		RunE: func(c *cobra.Command, args []string) error {
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

	cmd.Flags().StringVar(&lookupKey, "lookup-key", "", "lookup key identifier (required)")
	cmd.Flags().StringVar(&displayName, "display-name", "", "display name (required)")
	cmd.MarkFlagRequired("lookup-key")
	cmd.MarkFlagRequired("display-name")
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
		Args:  cobra.ExactArgs(1),
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
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
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
		Args:  cobra.ExactArgs(1),
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
		Args:  cobra.ExactArgs(1),
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
