package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

// Format is the output format type.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

// ColorDisabled is set to true when --no-color is passed or NO_COLOR env var is set.
var ColorDisabled bool

// PrettyJSON controls whether JSON output is indented.
// Defaults to true for TTY, false for pipes. --pretty overrides to true.
var PrettyJSON bool

func init() {
	if os.Getenv("NO_COLOR") != "" {
		ColorDisabled = true
	}
	PrettyJSON = IsTTY()
}

// IsTTY returns true if stdout is a terminal.
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// colorEnabled returns true if color output should be used.
func colorEnabled() bool {
	return IsTTY() && !ColorDisabled
}

// Print outputs data in the specified format.
func Print(format Format, data any, tableRenderer func(t table.Writer)) {
	switch format {
	case FormatJSON:
		printJSON(data)
	default:
		if tableRenderer != nil {
			printTable(tableRenderer)
		} else {
			printJSON(data)
		}
	}
}

func printJSON(data any) {
	enc := json.NewEncoder(os.Stdout)
	if PrettyJSON {
		enc.SetIndent("", "  ")
	}
	_ = enc.Encode(data)
}

func printTable(renderer func(t table.Writer)) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	renderer(t)
	t.Render()
}

// FormatTimestamp converts a millisecond epoch timestamp to a human-readable string.
func FormatTimestamp(ms int64) string {
	return time.UnixMilli(ms).Format("2006-01-02 15:04")
}

// Deref safely dereferences a string pointer, returning a fallback if nil.
func Deref(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

// ANSI color codes for terminal output.
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// ColorRed returns the red ANSI escape code if color is enabled, empty string otherwise.
func ColorRed() string {
	if colorEnabled() {
		return colorRed
	}
	return ""
}

// ColorReset returns the reset ANSI escape code if color is enabled, empty string otherwise.
func ColorReset() string {
	if colorEnabled() {
		return colorReset
	}
	return ""
}

// Success prints a success message to stderr (keeps stdout clean for piping).
func Success(msg string, args ...any) {
	if colorEnabled() {
		fmt.Fprintf(os.Stderr, colorGreen+"  "+msg+colorReset+"\n", args...)
	} else {
		fmt.Fprintf(os.Stderr, "  "+msg+"\n", args...)
	}
}

// Warn prints a warning message to stderr.
func Warn(msg string, args ...any) {
	if colorEnabled() {
		fmt.Fprintf(os.Stderr, colorYellow+"  Warning: "+msg+colorReset+"\n", args...)
	} else {
		fmt.Fprintf(os.Stderr, "  Warning: "+msg+"\n", args...)
	}
}

// Progress prints a progress indicator to stderr for bulk operations.
func Progress(current, total int, msg string, args ...any) {
	prefix := fmt.Sprintf("[%d/%d] ", current, total)
	fmt.Fprintf(os.Stderr, "  "+prefix+msg+"\n", args...)
}
