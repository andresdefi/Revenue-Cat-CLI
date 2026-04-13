package apps

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

func NewAppsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "apps",
		Aliases: []string{"app"},
		Short:   "Manage apps within a project",
		Long: `Manage platform apps within a RevenueCat project.

Each project can have multiple apps for different platforms (iOS, Android, Stripe, etc).
Products are created per-app using the store identifier from the respective platform.

Examples:
  rc apps list
  rc apps get app1a2b3c4
  rc apps create --name "My iOS App" --type app_store --bundle-id com.example.app
  rc apps delete app1a2b3c4`,
	}

	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(newGetCmd(projectID, outputFormat))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	root.AddCommand(newUpdateCmd(projectID, outputFormat))
	root.AddCommand(newDeleteCmd(projectID))
	root.AddCommand(newPublicKeysCmd(projectID, outputFormat))
	root.AddCommand(newStoreKitConfigCmd(projectID, outputFormat))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List apps in a project",
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			data, err := client.Get(fmt.Sprintf("/projects/%s/apps", url.PathEscape(pid)), nil)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.App]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Name", "Type", "Created"})
				for _, a := range resp.Items {
					t.AppendRow(table.Row{a.ID, a.Name, a.Type, output.FormatTimestamp(a.CreatedAt)})
				}
			})
			return nil
		},
	}
}

func newGetCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "get <app-id>",
		Short: "Get an app by ID",
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

			data, err := client.Get(fmt.Sprintf("/projects/%s/apps/%s", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}

			var app api.App
			if err := json.Unmarshal(data, &app); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, app, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", app.ID},
					{"Name", app.Name},
					{"Type", app.Type},
					{"Project ID", app.ProjectID},
					{"Created", output.FormatTimestamp(app.CreatedAt)},
				})
			})
			return nil
		},
	}
}

func newCreateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		name     string
		appType  string
		bundleID string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new app",
		Long: `Create a new platform app in a project.

Supported types: app_store, play_store, amazon, stripe, rc_billing,
roku, mac_app_store, paddle

Examples:
  rc apps create --name "iOS App" --type app_store --bundle-id com.example.app
  rc apps create --name "Android App" --type play_store --bundle-id com.example.app`,
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
				"name": name,
				"type": appType,
			}
			if bundleID != "" {
				switch appType {
				case "app_store", "mac_app_store":
					body["app_store"] = map[string]any{"bundle_id": bundleID}
				case "play_store":
					body["play_store"] = map[string]any{"package_name": bundleID}
				case "amazon":
					body["amazon"] = map[string]any{"package_name": bundleID}
				}
			}

			data, err := client.Post(fmt.Sprintf("/projects/%s/apps", url.PathEscape(pid)), body)
			if err != nil {
				return err
			}

			var app api.App
			if err := json.Unmarshal(data, &app); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, app, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", app.ID},
					{"Name", app.Name},
					{"Type", app.Type},
					{"Created", output.FormatTimestamp(app.CreatedAt)},
				})
			})
			output.Success("App created successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "app name (required)")
	cmd.Flags().StringVar(&appType, "type", "", "platform type: app_store, play_store, amazon, stripe, rc_billing, roku, mac_app_store, paddle (required)")
	cmd.Flags().StringVar(&bundleID, "bundle-id", "", "bundle ID / package name (for app_store, play_store, amazon)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("type")
	return cmd
}

func newUpdateCmd(projectID, outputFormat *string) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "update <app-id>",
		Short: "Update an app",
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
			if c.Flags().Changed("name") {
				body["name"] = name
			}

			data, err := client.Post(fmt.Sprintf("/projects/%s/apps/%s", url.PathEscape(pid), url.PathEscape(args[0])), body)
			if err != nil {
				return err
			}

			var app api.App
			if err := json.Unmarshal(data, &app); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, app, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", app.ID},
					{"Name", app.Name},
					{"Type", app.Type},
				})
			})
			output.Success("App updated")
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "new app name")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <app-id>",
		Short: "Delete an app",
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
			_, err = client.Delete(fmt.Sprintf("/projects/%s/apps/%s", url.PathEscape(pid), url.PathEscape(args[0])))
			if err != nil {
				return err
			}
			output.Success("App %s deleted", args[0])
			return nil
		},
	}
}

func newPublicKeysCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "public-keys <app-id>",
		Short: "List public API keys for an app",
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

			data, err := client.Get(fmt.Sprintf("/projects/%s/apps/%s/public_api_keys", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			if format == output.FormatJSON {
				var raw json.RawMessage
				if err := json.Unmarshal(data, &raw); err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}
				output.Print(format, raw, nil)
			} else {
				var resp api.ListResponse[api.PublicAPIKey]
				if err := json.Unmarshal(data, &resp); err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}
				output.Print(format, resp, func(t table.Writer) {
					t.AppendHeader(table.Row{"Key", "Name"})
					for _, k := range resp.Items {
						t.AppendRow(table.Row{k.Key, k.Name})
					}
				})
			}
			return nil
		},
	}
}

func newStoreKitConfigCmd(projectID, outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "storekit-config <app-id>",
		Short: "Get StoreKit configuration for an app",
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

			data, err := client.Get(fmt.Sprintf("/projects/%s/apps/%s/store_kit_config", url.PathEscape(pid), url.PathEscape(args[0])), nil)
			if err != nil {
				return err
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			if format == output.FormatJSON {
				var raw json.RawMessage
				if err := json.Unmarshal(data, &raw); err != nil {
					return fmt.Errorf("failed to parse response: %w", err)
				}
				output.Print(format, raw, nil)
			} else {
				// StoreKit config is opaque - just print as JSON since it's a config blob
				fmt.Println(string(data))
			}
			return nil
		},
	}
}
