package cmdutil

import (
	"fmt"

	"github.com/andresdefi/rc/internal/config"
	"github.com/andresdefi/rc/internal/output"
)

// ActiveProfile is set by the root command's --profile flag.
// Empty string means use the config's current_profile.
var ActiveProfile string

// ResolveProfile returns the effective profile name.
func ResolveProfile() string {
	if ActiveProfile != "" {
		return ActiveProfile
	}
	cfg, err := config.Load()
	if err != nil {
		return config.DefaultProfileName
	}
	if cfg.CurrentProfile != "" {
		return cfg.CurrentProfile
	}
	return config.DefaultProfileName
}

// ResolveProject returns the project ID from the flag, the active profile, or an error.
func ResolveProject(flagValue *string) (string, error) {
	if flagValue != nil && *flagValue != "" {
		return *flagValue, nil
	}

	cfg, err := config.Load()
	if err == nil {
		profile := ResolveProfile()
		p := cfg.GetProfile(profile)
		if p != nil && p.ProjectID != "" {
			return p.ProjectID, nil
		}
	}

	return "", fmt.Errorf("no project specified - use --project flag or run `rc projects set-default <project-id>`")
}

// GetOutputFormat returns the resolved output format.
// If the user explicitly set --output, use that. Otherwise, default to
// table for TTY and JSON for pipes (so `rc products list | jq` just works).
func GetOutputFormat(flag *string) output.Format {
	if flag != nil {
		switch *flag {
		case "json":
			return output.FormatJSON
		case "table":
			return output.FormatTable
		}
	}
	// Auto-detect: table for terminal, JSON for pipes
	if output.IsTTY() {
		return output.FormatTable
	}
	return output.FormatJSON
}
