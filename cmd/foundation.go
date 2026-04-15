package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/andresdefi/rc/internal/api"
	internalAuth "github.com/andresdefi/rc/internal/auth"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/config"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newInitCmd(projectID *string) *cobra.Command {
	var (
		profileName string
		current     bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize rc configuration for a profile",
		Long: `Initialize rc configuration for a profile.

This creates ~/.rc/config.toml when needed, records the default project for
the selected profile, and can make that profile current. It does not prompt for
an API key; run rc auth login before or after init.`,
		Example: `  # Initialize the default profile with a project
  rc init --project proj1a2b3c4d5

  # Initialize a staging profile and make it current
  rc init --profile-name staging --project proj_staging --current

  # Log in and validate after initialization
  rc auth login --profile staging
  rc doctor --profile staging`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := profileName
			if name == "" {
				name = cmdutil.ResolveProfile()
			}
			if name == "" {
				name = config.DefaultProfileName
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			p := cfg.GetProfile(name)
			if p == nil {
				p = &config.Profile{}
			}
			if projectID != nil && *projectID != "" {
				p.ProjectID = *projectID
			}
			cfg.SetProfile(name, p)
			if current {
				cfg.CurrentProfile = name
			}
			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			path, _ := config.Path()
			output.Success("Initialized profile %s in %s", name, path)
			if p.ProjectID == "" {
				output.Warn("No default project set; pass --project or run `rc projects set-default <project-id>`")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&profileName, "profile-name", "", "profile to initialize (defaults to active profile)")
	cmd.Flags().BoolVar(&current, "current", true, "set the initialized profile as current")
	return cmd
}

func newDoctorCmd(projectID *string) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check rc configuration, authentication, and API connectivity",
		Example: `  # Check the active profile
  rc doctor

  # Check a specific profile
  rc doctor --profile staging

  # Check a specific project override
  rc doctor --project proj1a2b3c4d5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := runDoctor(projectID)
			printDoctor(report)
			return nil
		},
	}
}

func newWhoamiCmd(outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the active profile, auth source, and default project",
		Example: `  # Show current identity context
  rc whoami

  # Script-friendly output
  rc whoami --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := cmdutil.ResolveProfile()
			cfg, _ := config.Load()
			var project string
			if cfg != nil {
				if p := cfg.GetProfile(profile); p != nil {
					project = p.ProjectID
				}
			}

			token, err := internalAuth.GetToken(profile)
			status := map[string]any{
				"profile":         profile,
				"default_project": project,
				"logged_in":       err == nil,
				"token_source":    "",
				"token":           "",
			}
			if err == nil {
				status["token_source"] = internalAuth.TokenSource(profile)
				status["token"] = internalAuth.MaskToken(token)
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, status, func(t table.Writer) {
				t.AppendHeader(table.Row{"Field", "Value"})
				t.AppendRows([]table.Row{
					{"Profile", status["profile"]},
					{"Logged in", status["logged_in"]},
					{"Token", status["token"]},
					{"Token source", status["token_source"]},
					{"Default project", status["default_project"]},
				})
			})
			return nil
		},
	}
}

type doctorReport struct {
	Profile        string `json:"profile"`
	ConfigPath     string `json:"config_path"`
	ConfigOK       bool   `json:"config_ok"`
	ConfigMessage  string `json:"config_message,omitempty"`
	TokenOK        bool   `json:"token_ok"`
	TokenMessage   string `json:"token_message,omitempty"`
	ProjectID      string `json:"project_id,omitempty"`
	ProjectOK      bool   `json:"project_ok"`
	ProjectMessage string `json:"project_message,omitempty"`
	APIOK          bool   `json:"api_ok"`
	APIMessage     string `json:"api_message,omitempty"`
}

func runDoctor(projectID *string) doctorReport {
	profile := cmdutil.ResolveProfile()
	path, _ := config.Path()
	report := doctorReport{Profile: profile, ConfigPath: path}

	_, err := config.Load()
	if err != nil {
		report.ConfigMessage = err.Error()
		return report
	}
	report.ConfigOK = true
	report.ConfigMessage = "OK"

	token, err := internalAuth.GetToken(profile)
	if err != nil {
		report.TokenMessage = "not logged in - run `rc auth login`"
		return report
	}
	report.TokenOK = true
	report.TokenMessage = fmt.Sprintf("OK (%s)", internalAuth.TokenSource(profile))

	pid, err := cmdutil.ResolveProject(projectID)
	if err != nil {
		report.ProjectMessage = err.Error()
	} else {
		report.ProjectID = pid
		report.ProjectOK = true
		report.ProjectMessage = "OK"
	}

	client := api.NewClientWithToken(token)
	data, err := client.Get("/projects", nil)
	if err != nil {
		report.APIMessage = err.Error()
		return report
	}

	var resp api.ListResponse[api.Project]
	if err := json.Unmarshal(data, &resp); err != nil {
		report.APIMessage = "could not parse /projects response"
		return report
	}
	report.APIOK = true
	report.APIMessage = fmt.Sprintf("OK (found %d projects)", len(resp.Items))
	return report
}

func printDoctor(report doctorReport) {
	fmt.Fprintf(os.Stderr, "Profile:     %s\n", report.Profile)
	fmt.Fprintf(os.Stderr, "Config:      %s\n", formatCheck(report.ConfigOK, report.ConfigMessage))
	fmt.Fprintf(os.Stderr, "Token:       %s\n", formatCheck(report.TokenOK, report.TokenMessage))
	fmt.Fprintf(os.Stderr, "Project:     %s\n", formatCheck(report.ProjectOK, report.ProjectMessage))
	fmt.Fprintf(os.Stderr, "API access:  %s\n", formatCheck(report.APIOK, report.APIMessage))
}

func formatCheck(ok bool, message string) string {
	if message == "" {
		if ok {
			return "OK"
		}
		return "not checked"
	}
	if ok {
		return message
	}
	return "failed - " + message
}
