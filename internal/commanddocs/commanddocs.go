package commanddocs

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/andresdefi/rc/internal/cmdutil"
)

// Generate renders a single-file Markdown command reference from a Cobra tree.
func Generate(root *cobra.Command) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("# rc Command Reference\n\n")
	buf.WriteString("Generated from Cobra command definitions. Do not edit by hand.\n\n")
	writeCommand(&buf, root, 2)
	return buf.Bytes(), nil
}

func writeCommand(buf *bytes.Buffer, cmd *cobra.Command, level int) {
	if cmd.Hidden {
		return
	}

	fmt.Fprintf(buf, "%s %s\n\n", strings.Repeat("#", level), commandPath(cmd))
	if cmd.Short != "" {
		fmt.Fprintf(buf, "%s\n\n", cmd.Short)
	}
	if stability := cmd.Annotations[cmdutil.StabilityKey]; stability != "" && stability != cmdutil.StabilityStable {
		fmt.Fprintf(buf, "**Stability:** `%s`\n\n", stability)
	}
	if cmd.Long != "" && cmd.Long != cmd.Short {
		fmt.Fprintf(buf, "%s\n\n", strings.TrimSpace(cmd.Long))
	}

	if cmd.HasAvailableLocalFlags() {
		buf.WriteString("**Flags**\n\n")
		cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
			if flag.Hidden {
				return
			}
			name := "--" + flag.Name
			if flag.Shorthand != "" {
				name = "-" + flag.Shorthand + ", " + name
			}
			def := ""
			if flag.DefValue != "" {
				def = fmt.Sprintf(" Default: `%s`.", flag.DefValue)
			}
			fmt.Fprintf(buf, "- `%s`: %s%s\n", name, flag.Usage, def)
		})
		buf.WriteString("\n")
	}

	if cmd.Example != "" {
		buf.WriteString("**Examples**\n\n")
		buf.WriteString("```bash\n")
		buf.WriteString(strings.TrimSpace(cmd.Example))
		buf.WriteString("\n```\n\n")
	}

	children := cmd.Commands()
	for _, child := range children {
		if !child.Hidden {
			writeCommand(buf, child, level+1)
		}
	}
}

func commandPath(cmd *cobra.Command) string {
	if cmd.HasParent() {
		return cmd.CommandPath()
	}
	return cmd.Use
}
