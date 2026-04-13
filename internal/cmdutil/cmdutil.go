package cmdutil

import (
	"fmt"

	"github.com/andresdefi/rc/internal/config"
	"github.com/andresdefi/rc/internal/output"
)

// ResolveProject returns the project ID from the flag or config.
func ResolveProject(flagValue *string) (string, error) {
	if flagValue != nil && *flagValue != "" {
		return *flagValue, nil
	}

	cfg, err := config.Load()
	if err == nil && cfg.ProjectID != "" {
		return cfg.ProjectID, nil
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
