package apps

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/completions"
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

	c := completions.AppIDs(projectID)
	root.AddCommand(newListCmd(projectID, outputFormat))
	root.AddCommand(completions.WithCompletion(newGetCmd(projectID, outputFormat), c))
	root.AddCommand(newCreateCmd(projectID, outputFormat))
	root.AddCommand(completions.WithCompletion(newUpdateCmd(projectID, outputFormat), c))
	root.AddCommand(completions.WithCompletion(newDeleteCmd(projectID), c))
	root.AddCommand(newPublicKeysCmd(projectID, outputFormat))
	root.AddCommand(newStoreKitConfigCmd(projectID, outputFormat))
	return root
}

func newListCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		fetchAll bool
		limit    int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List apps in a project",
		Example: `  # List all apps
  rc apps list

  # List with JSON output
  rc apps list -o json

  # Fetch all pages
  rc apps list --all`,
		RunE: func(c *cobra.Command, args []string) error {
			pid, err := cmdutil.ResolveProject(projectID)
			if err != nil {
				return err
			}
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			path := fmt.Sprintf("/projects/%s/apps", url.PathEscape(pid))
			query := url.Values{}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}

			if fetchAll {
				items, err := api.PaginateAll[api.App](client, path, query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Name", "Type", "Created"})
					for _, a := range items {
						t.AppendRow(table.Row{a.ID, a.Name, a.Type, output.FormatTimestamp(a.CreatedAt)})
					}
					t.AppendFooter(table.Row{"", "", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}

			data, err := client.Get(path, query)
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
		Use:   "get <app-id>",
		Short: "Get an app by ID",
		Example: `  # Get app details
  rc apps get app1a2b3c4d5

  # Get as JSON
  rc apps get app1a2b3c4d5 -o json`,
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
roku, mac_app_store, paddle.

Required flags are prompted interactively when running in a terminal and not
provided on the command line.`,
		Example: `  # Create an iOS app
  rc apps create --name "iOS App" --type app_store --bundle-id com.example.app

  # Create an Android app
  rc apps create --name "Android App" --type play_store --bundle-id com.example.app

  # Create a Stripe app
  rc apps create --name "Web Payments" --type stripe

  # Interactive mode (prompts for missing fields)
  rc apps create`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.PromptIfEmpty(&name, "App name", "iOS App"); err != nil {
				return err
			}
			if err := cmdutil.PromptSelect(&appType, "Platform type", []string{"app_store", "play_store", "amazon", "stripe", "rc_billing", "roku", "mac_app_store", "paddle"}); err != nil {
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
	return cmd
}

func newUpdateCmd(projectID, outputFormat *string) *cobra.Command {
	var (
		name                string
		sharedSecret        string
		subscriptionKeyFile string
		subscriptionKeyID   string
		subscriptionIssuer  string
		connectKeyFile      string
		connectKeyID        string
		connectIssuer       string
		connectVendorNumber string
		serviceAccountFile  string
	)

	cmd := &cobra.Command{
		Use:   "update <app-id>",
		Short: "Update an app",
		Example: `  # Rename an app
  rc apps update app1a2b3c4d5 --name "My Renamed App"

  # Configure App Store shared secret
  rc apps update app1a2b3c4d5 --shared-secret 1234567890abcdef1234567890abcdef

  # Configure App Store in-app purchase key
  rc apps update app1a2b3c4d5 \
    --subscription-key-file ./SubscriptionKey_ABC123.p8 \
    --subscription-key-id ABC123 \
    --subscription-key-issuer 5a049d62-1b9b-453c-b605-1988189d8129

  # Configure App Store Connect API credentials
  rc apps update app1a2b3c4d5 \
    --app-store-connect-api-key-file ./AuthKey_ABC123.p8 \
    --app-store-connect-api-key-id ABC123 \
    --app-store-connect-api-key-issuer 5a049d62-1b9b-453c-b605-1988189d8129 \
    --app-store-connect-vendor-number 12345678`,
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
			if c.Flags().Changed("name") {
				body["name"] = name
			}
			if c.Flags().Changed("service-account-file") {
				return fmt.Errorf("RevenueCat API v2 does not document a Play Store service-account credential field for app updates; no request sent")
			}

			appStore := map[string]any{}
			if c.Flags().Changed("shared-secret") {
				appStore["shared_secret"] = sharedSecret
			}
			if c.Flags().Changed("subscription-key-file") {
				contents, err := os.ReadFile(subscriptionKeyFile)
				if err != nil {
					return fmt.Errorf("read subscription key file: %w", err)
				}
				appStore["subscription_private_key"] = string(contents)
			}
			if c.Flags().Changed("subscription-key-id") {
				appStore["subscription_key_id"] = subscriptionKeyID
			}
			if c.Flags().Changed("subscription-key-issuer") {
				appStore["subscription_key_issuer"] = subscriptionIssuer
			}
			if c.Flags().Changed("app-store-connect-api-key-file") {
				contents, err := os.ReadFile(connectKeyFile)
				if err != nil {
					return fmt.Errorf("read App Store Connect API key file: %w", err)
				}
				appStore["app_store_connect_api_key"] = string(contents)
			}
			if c.Flags().Changed("app-store-connect-api-key-id") {
				appStore["app_store_connect_api_key_id"] = connectKeyID
			}
			if c.Flags().Changed("app-store-connect-api-key-issuer") {
				appStore["app_store_connect_api_key_issuer"] = connectIssuer
			}
			if c.Flags().Changed("app-store-connect-vendor-number") {
				appStore["app_store_connect_vendor_number"] = connectVendorNumber
			}
			if len(appStore) > 0 {
				body["app_store"] = appStore
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
	cmd.Flags().StringVar(&sharedSecret, "shared-secret", "", "App Store shared secret")
	cmd.Flags().StringVar(&subscriptionKeyFile, "subscription-key-file", "", "path to App Store in-app purchase key .p8 file")
	cmd.Flags().StringVar(&subscriptionKeyID, "subscription-key-id", "", "App Store in-app purchase key ID")
	cmd.Flags().StringVar(&subscriptionIssuer, "subscription-key-issuer", "", "App Store in-app purchase key issuer ID")
	cmd.Flags().StringVar(&connectKeyFile, "app-store-connect-api-key-file", "", "path to App Store Connect API key .p8 file")
	cmd.Flags().StringVar(&connectKeyID, "app-store-connect-api-key-id", "", "App Store Connect API key ID")
	cmd.Flags().StringVar(&connectIssuer, "app-store-connect-api-key-issuer", "", "App Store Connect API key issuer ID")
	cmd.Flags().StringVar(&connectVendorNumber, "app-store-connect-vendor-number", "", "App Store Connect vendor number")
	cmd.Flags().StringVar(&serviceAccountFile, "service-account-file", "", "path to Google Play service account JSON file (not supported by RevenueCat API v2 app update)")
	return cmd
}

func newDeleteCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <app-id>",
		Short: "Delete an app",
		Example: `  # Delete an app
  rc apps delete app1a2b3c4d5`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.ConfirmDestructive("Delete", "app", args[0]); err != nil {
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
		Example: `  # List public keys for an app
  rc apps public-keys app1a2b3c4d5

  # Get as JSON
  rc apps public-keys app1a2b3c4d5 -o json`,
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
		Example: `  # Get StoreKit config for an iOS app
  rc apps storekit-config app1a2b3c4d5

  # Get as JSON
  rc apps storekit-config app1a2b3c4d5 -o json`,
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
