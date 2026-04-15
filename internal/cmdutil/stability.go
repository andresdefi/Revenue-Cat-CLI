package cmdutil

import "github.com/spf13/cobra"

// Stability annotation key and values for Cobra commands.
const (
	StabilityKey = "stability"

	StabilityStable       = "stable"
	StabilityBeta         = "beta"
	StabilityExperimental = "experimental"
	StabilityDeprecated   = "deprecated"
)

// stabilityLabels maps stability values to display labels.
var stabilityLabels = map[string]string{
	StabilityBeta:         "[beta] ",
	StabilityExperimental: "[experimental] ",
	StabilityDeprecated:   "[deprecated] ",
}

// MarkBeta marks a command as beta and prepends a label to its Short description.
func MarkBeta(cmd *cobra.Command) {
	markStability(cmd, StabilityBeta)
}

// MarkExperimental marks a command as experimental.
func MarkExperimental(cmd *cobra.Command) {
	markStability(cmd, StabilityExperimental)
}

// MarkDeprecated marks a command as deprecated.
func MarkDeprecated(cmd *cobra.Command) {
	markStability(cmd, StabilityDeprecated)
}

func markStability(cmd *cobra.Command, level string) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations[StabilityKey] = level
	if label, ok := stabilityLabels[level]; ok {
		cmd.Short = label + cmd.Short
	}
}
