package auth

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/andresdefi/rc/internal/api"
	internalAuth "github.com/andresdefi/rc/internal/auth"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/output"
	"github.com/spf13/cobra"
)

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long: `Log in, check status, or log out of the RevenueCat API.

Working with multiple projects:
RevenueCat API v2 secret keys are project-scoped, so each project needs its
own profile. Add keys with:
  rc auth add-project --key sk_xxx --name my-project

Then query every stored profile at once:
  rc projects list --all-profiles`,
	}

	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newAddProjectCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newDoctorCmd())
	cmd.AddCommand(newValidateCmd())
	return cmd
}

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a RevenueCat API v2 secret key",
		Long: `Authenticate with your RevenueCat API v2 secret key.

You can create a v2 secret key in the RevenueCat dashboard:
  Project Settings > API Keys > + New Secret API Key

The key will be stored in your system keychain (with config file fallback).
Keys are prefixed with sk_ and must have v2 API permissions.`,
		Example: `  # Log in with the default profile
  rc auth login

  # Log in with a specific profile
  rc auth login --profile staging

  # Check the profile after login
  rc auth login --profile production
  rc auth status --profile production

  # Verify API access after saving a key
  rc auth login
  rc auth doctor

  # Switch between profiles for project work
  rc auth login --profile staging
  rc projects list --profile staging`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := cmdutil.ResolveProfile()

			fmt.Print("Enter your RevenueCat API v2 secret key: ")
			reader := bufio.NewReader(os.Stdin)
			token, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			token = strings.TrimSpace(token)

			if err := validateAPIKey(token); err != nil {
				return err
			}

			if err := internalAuth.SaveToken(profile, token); err != nil {
				return fmt.Errorf("failed to save API key: %w", err)
			}

			output.Success("Logged in successfully [profile: %s] (stored in %s)", profile, internalAuth.TokenSource(profile))
			output.Next("rc projects list")
			return nil
		},
	}
}

func newAddProjectCmd() *cobra.Command {
	var key string
	var name string

	cmd := &cobra.Command{
		Use:   "add-project",
		Short: "Add a project-scoped API key as a named profile",
		Long: `Add a project-scoped RevenueCat API v2 key as a profile.

RevenueCat API v2 secret keys are scoped to one project. Use one profile per
project, then query all configured project keys with:
  rc projects list --all-profiles`,
		Example: `  # Add a key under an explicit profile name
  rc auth add-project --key sk_xxx --name impostor

  # Infer the profile name from the project visible to the key
  rc auth add-project --key sk_xxx

  # List projects visible across every stored profile
  rc projects list --all-profiles`,
		RunE: func(cmd *cobra.Command, args []string) error {
			key = strings.TrimSpace(key)
			name = strings.TrimSpace(name)
			if err := validateAPIKey(key); err != nil {
				return err
			}

			projects, err := fetchProjectsForKey(key)
			if err != nil {
				return fmt.Errorf("failed to validate API key: %w", err)
			}
			if len(projects) == 0 {
				return fmt.Errorf("API key is valid but did not return any projects")
			}

			profileName, err := resolveAddProjectProfileName(name, projects)
			if err != nil {
				return err
			}

			if err := internalAuth.SaveToken(profileName, key); err != nil {
				return fmt.Errorf("failed to save API key: %w", err)
			}

			project := projects[0]
			output.Success("Added project '%s' (%s) under profile '%s'", project.Name, project.ID, profileName)
			output.Next("rc projects list --profile %s", profileName)
			return nil
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "RevenueCat API v2 secret key (required)")
	cmd.Flags().StringVar(&name, "name", "", "profile name to store the key under")
	return cmd
}

func validateAPIKey(token string) error {
	if token == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	if !strings.HasPrefix(token, "sk_") && !strings.HasPrefix(token, "atk_") {
		return fmt.Errorf("invalid API key prefix: keys must start with 'sk_' (secret key) or 'atk_' (OAuth token)")
	}
	if len(token) < 10 {
		return fmt.Errorf("API key is too short - check that you copied the full key")
	}
	return nil
}

func fetchProjectsForKey(key string) ([]api.Project, error) {
	client := api.NewClientWithToken(key)
	data, err := client.Get("/projects", nil)
	if err != nil {
		return nil, err
	}
	var resp api.ListResponse[api.Project]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return resp.Items, nil
}

func resolveAddProjectProfileName(name string, projects []api.Project) (string, error) {
	if name != "" {
		return name, nil
	}
	if len(projects) > 1 {
		return "", fmt.Errorf("API key returned multiple projects; pass --name to choose a profile name")
	}
	return inferProfileName(projects[0]), nil
}

func inferProfileName(project api.Project) string {
	name := strings.ToLower(project.Name)
	var b strings.Builder
	lastWasSeparator := false
	for _, r := range name {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastWasSeparator = false
		case !lastWasSeparator && b.Len() > 0:
			b.WriteByte('-')
			lastWasSeparator = true
		}
	}
	profile := strings.Trim(b.String(), "-")
	if profile == "" {
		return project.ID
	}
	return profile
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		Example: `  # Check auth status for the default profile
  rc auth status

  # Check auth status for a specific profile
  rc auth status --profile production

  # Run diagnostics after checking status
  rc auth status
  rc auth doctor

  # Confirm a profile before listing projects
  rc auth status --profile staging
  rc projects list --profile staging

  # Use in scripts before a workflow
  rc auth status --profile production >/dev/null && rc products list --profile production --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := cmdutil.ResolveProfile()

			token, err := internalAuth.GetToken(profile)
			if err != nil {
				fmt.Printf("Not logged in [profile: %s]\n", profile)
				fmt.Println("Run `rc auth login` to authenticate")
				printStoredProfileCount()
				return nil
			}

			fmt.Printf("Logged in [profile: %s]\n", profile)
			fmt.Printf("  Key:     %s\n", internalAuth.MaskToken(token))
			fmt.Printf("  Stored:  %s\n", internalAuth.TokenSource(profile))
			printStoredProfileCount()
			return nil
		},
	}
}

func printStoredProfileCount() {
	count, err := internalAuth.StoredProfileCount()
	if err == nil && count > 1 {
		fmt.Printf("  Profiles: %d stored\n", count)
	}
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored API key",
		Example: `  # Log out of the default profile
  rc auth logout

  # Log out of a specific profile
  rc auth logout --profile staging`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := cmdutil.ResolveProfile()

			if err := internalAuth.DeleteToken(profile); err != nil {
				return fmt.Errorf("failed to remove API key: %w", err)
			}
			output.Success("Logged out successfully [profile: %s]", profile)
			return nil
		},
	}
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check authentication health and API connectivity",
		Example: `  # Run auth diagnostics
  rc auth doctor

  # Check a specific profile
  rc auth doctor --profile production

  # Validate auth before listing projects
  rc auth doctor
  rc projects list

  # Diagnose a staging profile
  rc auth status --profile staging
  rc auth doctor --profile staging

  # Use in scripts before a release check
  rc auth doctor --profile production >/dev/null && rc products list --profile production --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := cmdutil.ResolveProfile()
			fmt.Fprintf(os.Stderr, "Profile:     %s\n", profile)

			token, err := internalAuth.GetToken(profile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Token:       not found\n")
				fmt.Fprintf(os.Stderr, "API access:  failed - run `rc auth login` to authenticate\n")
				return nil
			}

			source := internalAuth.TokenSource(profile)
			fmt.Fprintf(os.Stderr, "Token:       %s (stored in %s)\n", internalAuth.MaskToken(token), source)

			client := api.NewClientWithToken(token)
			data, err := client.Get("/projects", nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "API access:  failed - %v\n", err)
				return nil
			}

			var resp api.ListResponse[api.Project]
			if err := json.Unmarshal(data, &resp); err != nil {
				fmt.Fprintf(os.Stderr, "API access:  failed - could not parse response\n")
				return nil
			}

			fmt.Fprintf(os.Stderr, "API access:  OK (found %d projects)\n", len(resp.Items))
			return nil
		},
	}
}

func newValidateCmd() *cobra.Command {
	cmd := newDoctorCmd()
	cmd.Use = "validate"
	cmd.Short = "Validate authentication and API connectivity"
	cmd.Example = `  # Validate the active profile
  rc auth validate

  # Validate a specific profile
  rc auth validate --profile production`
	return cmd
}
