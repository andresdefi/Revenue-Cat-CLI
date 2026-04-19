package projects

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/completions"
	"github.com/andresdefi/rc/internal/config"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewProjectsCmd(projectID, outputFormat *string) *cobra.Command {
	root := &cobra.Command{
		Use:     "projects",
		Aliases: []string{"project", "proj"},
		Short:   "Manage RevenueCat projects",
	}

	root.AddCommand(newListCmd(outputFormat))
	root.AddCommand(newCreateCmd(outputFormat))
	root.AddCommand(newDoctorCmd(projectID, outputFormat))
	root.AddCommand(completions.WithCompletion(newSetDefaultCmd(), completions.ProjectIDs()))
	return root
}

func newListCmd(outputFormat *string) *cobra.Command {
	var fetchAll bool
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all projects",
		Example: `  # List all projects
  rc projects list

  # List with JSON output
  rc projects list -o json

  # Fetch all pages
  rc projects list --all`,
		RunE: func(c *cobra.Command, args []string) error {
			client, err := api.NewClient()
			if err != nil {
				return err
			}

			query := url.Values{}
			if limit > 0 {
				query.Set("limit", fmt.Sprintf("%d", limit))
			}

			if fetchAll {
				items, err := api.PaginateAll[api.Project](client, "/projects", query)
				if err != nil {
					return err
				}
				format := cmdutil.GetOutputFormat(outputFormat)
				output.Print(format, items, func(t table.Writer) {
					t.AppendHeader(table.Row{"ID", "Name", "Created"})
					for _, p := range items {
						t.AppendRow(table.Row{p.ID, p.Name, output.FormatTimestamp(p.CreatedAt)})
					}
					t.AppendFooter(table.Row{"", "", fmt.Sprintf("%d total", len(items))})
				})
				return nil
			}

			data, err := client.Get("/projects", query)
			if err != nil {
				return err
			}

			var resp api.ListResponse[api.Project]
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, resp, func(t table.Writer) {
				t.AppendHeader(table.Row{"ID", "Name", "Created"})
				for _, p := range resp.Items {
					t.AppendRow(table.Row{p.ID, p.Name, output.FormatTimestamp(p.CreatedAt)})
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

func newCreateCmd(outputFormat *string) *cobra.Command {
	var name string

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long: `Create a new project. Required flags are prompted interactively when
running in a terminal and not provided on the command line.`,
		Example: `  # Create a new project
  rc projects create --name "My App"

  # Create and output as JSON
  rc projects create --name "My App" -o json

  # Interactive mode (prompts for missing fields)
  rc projects create`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := cmdutil.PromptIfEmpty(&name, "Project name", "My App"); err != nil {
				return err
			}

			client, err := api.NewClient()
			if err != nil {
				return err
			}

			data, err := client.Post("/projects", map[string]any{"name": name})
			if err != nil {
				return err
			}

			var project api.Project
			if err := json.Unmarshal(data, &project); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, project, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"ID", project.ID},
					{"Name", project.Name},
					{"Created", output.FormatTimestamp(project.CreatedAt)},
				})
			})
			output.Success("Project created successfully")
			output.Next("rc projects set-default %s", project.ID)
			return nil
		},
	}

	createCmd.Flags().StringVar(&name, "name", "", "project name (required)")
	return createCmd
}

func newSetDefaultCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-default <project-id>",
		Short: "Set the default project for all commands",
		Example: `  # Set the default project
  rc projects set-default proj1a2b3c4d5

  # Set default for a specific profile
  rc projects set-default proj1a2b3c4d5 --profile staging`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			profile := cmdutil.ResolveProfile()
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			p := cfg.GetProfile(profile)
			if p == nil {
				p = &config.Profile{}
			}
			p.ProjectID = args[0]
			cfg.SetProfile(profile, p)
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
			output.Success("Default project set to %s [profile: %s]", args[0], profile)
			output.Next("rc apps list")
			return nil
		},
	}
}
