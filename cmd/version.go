package cmd

import (
	"fmt"

	"github.com/andresdefi/rc/internal/version"
	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of rc",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("rc %s (commit: %s, built: %s)\n", version.Version, version.Commit, version.Date)
		},
	}
}
