package config

import (
	"fmt"
	"sort"

	"github.com/andresdefi/rc/internal/cmdutil"
	internalConfig "github.com/andresdefi/rc/internal/config"
	"github.com/andresdefi/rc/internal/output"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewConfigCmd(outputFormat *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect rc configuration",
		Long:  "Inspect local rc configuration, profiles, and default project settings.",
	}

	cmd.AddCommand(newProfilesCmd(outputFormat))
	return cmd
}

type profileRow struct {
	Name      string `json:"name"`
	Current   bool   `json:"current"`
	ProjectID string `json:"project_id,omitempty"`
	HasAPIKey bool   `json:"has_config_api_key"`
}

func newProfilesCmd(outputFormat *string) *cobra.Command {
	return &cobra.Command{
		Use:   "profiles",
		Short: "List configured profiles",
		Example: `  # List profiles
  rc config profiles

  # Script-friendly profile list
  rc config profiles --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := internalConfig.Load()
			if err != nil {
				return err
			}
			names := make([]string, 0, len(cfg.Profiles))
			for name := range cfg.Profiles {
				names = append(names, name)
			}
			sort.Strings(names)

			rows := make([]profileRow, 0, len(names))
			for _, name := range names {
				p := cfg.Profiles[name]
				if p == nil {
					p = &internalConfig.Profile{}
				}
				rows = append(rows, profileRow{
					Name:      name,
					Current:   name == cfg.CurrentProfile,
					ProjectID: p.ProjectID,
					HasAPIKey: p.APIKey != "",
				})
			}

			format := cmdutil.GetOutputFormat(outputFormat)
			output.Print(format, rows, func(t table.Writer) {
				t.AppendHeader(table.Row{"Profile", "Current", "Default Project", "Config API Key"})
				for _, row := range rows {
					current := ""
					if row.Current {
						current = "yes"
					}
					t.AppendRow(table.Row{row.Name, current, row.ProjectID, fmt.Sprintf("%t", row.HasAPIKey)})
				}
			})
			return nil
		},
	}
}
