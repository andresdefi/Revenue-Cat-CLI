package offerings

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewOfferingsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "offerings",
		Aliases: []string{"offering", "off"},
		Short:   "Manage offerings",
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
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
					t.AppendRow(table.Row{
						o.ID,
						o.LookupKey,
						o.DisplayName,
						current,
						o.State,
						pkgCount,
						output.FormatTimestamp(o.CreatedAt),
					})
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

	createCmd := &cobra.Command{
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

			body := map[string]any{
				"lookup_key":   lookupKey,
				"display_name": displayName,
			}

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

	createCmd.Flags().StringVar(&lookupKey, "lookup-key", "", "lookup key identifier (required)")
	createCmd.Flags().StringVar(&displayName, "display-name", "", "display name (required)")
	createCmd.MarkFlagRequired("lookup-key")
	createCmd.MarkFlagRequired("display-name")
	return createCmd
}
