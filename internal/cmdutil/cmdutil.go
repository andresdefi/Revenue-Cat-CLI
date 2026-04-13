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
func GetOutputFormat(flag *string) output.Format {
	if flag != nil && *flag == "json" {
		return output.FormatJSON
	}
	return output.FormatTable
}
