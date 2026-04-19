package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

// Format is the output format type.
type Format string

const (
	FormatTable    Format = "table"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
)

// ValidFormats lists all accepted --output values.
var ValidFormats = []string{"table", "json", "markdown", "md"}

// ColorDisabled is set to true when --no-color is passed or NO_COLOR env var is set.
var ColorDisabled bool

// PrettyJSON controls whether JSON output is indented.
// Defaults to true for TTY, false for pipes. --pretty overrides to true.
var PrettyJSON bool

// Quiet suppresses non-essential output (Success, Warn, Progress).
// Only data output and errors pass through.
var Quiet bool

// HintsDisabled suppresses post-mutation next-step hints.
var HintsDisabled bool

// FieldsFilter is the comma-separated list of fields to include in JSON output.
// Set by cmdutil.FieldsFlag from --fields flag.
var FieldsFilter string

// DefaultFieldsPreset is the active command's preset for --fields default.
var DefaultFieldsPreset string

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
// If fields are specified via --fields, JSON output is filtered to those fields.
func Print(format Format, data any, tableRenderer func(t table.Writer)) {
	switch format {
	case FormatJSON:
		printJSON(filterFields(data))
	case FormatMarkdown:
		if tableRenderer != nil {
			printMarkdown(tableRenderer)
		} else {
			printJSON(filterFields(data))
		}
	default:
		if tableRenderer != nil {
			printTable(tableRenderer)
		} else {
			printJSON(filterFields(data))
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

func printMarkdown(renderer func(t table.Writer)) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	renderer(t)
	_, _ = fmt.Fprintln(os.Stdout, t.RenderMarkdown())
}

// filterFields reduces JSON data to only the requested fields.
// Works on objects, list envelopes, and raw slices.
func filterFields(data any) any {
	fields, ok := resolveFieldsFilter()
	if !ok {
		return data
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return data
	}

	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return data
	}
	return filterValue(value, fields)
}

func resolveFieldsFilter() ([]string, bool) {
	if FieldsFilter == "" {
		return nil, false
	}
	filter := FieldsFilter
	if filter == "default" {
		filter = DefaultFieldsPreset
		if filter == "" {
			if LogLevel >= LogLevelWarn {
				Warn("no default preset for this command, returning full output")
			}
			return nil, false
		}
	}
	fields := strings.Split(filter, ",")
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if trimmed := strings.TrimSpace(field); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result, len(result) > 0
}

func filterValue(value any, fields []string) any {
	obj, ok := value.(map[string]any)
	if !ok {
		if arr, ok := value.([]any); ok {
			return filterItems(arr, fields)
		}
		return value
	}
	if items, ok := obj["items"]; ok {
		if arr, ok := items.([]any); ok {
			obj["items"] = filterItems(arr, fields)
			return obj
		}
	}
	return pickFields(obj, fields)
}

func filterItems(items []any, fields []string) []any {
	filtered := make([]any, 0, len(items))
	for _, item := range items {
		if m, ok := item.(map[string]any); ok {
			filtered = append(filtered, pickFields(m, fields))
		}
	}
	return filtered
}

func pickFields(obj map[string]any, fields []string) map[string]any {
	result := make(map[string]any, len(fields))
	for _, f := range fields {
		if v, ok := obj[f]; ok {
			result[f] = v
		}
	}
	return result
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
	if Quiet {
		return
	}
	if colorEnabled() {
		fmt.Fprintf(os.Stderr, colorGreen+"  "+msg+colorReset+"\n", args...)
	} else {
		fmt.Fprintf(os.Stderr, "  "+msg+"\n", args...)
	}
}

// Next prints a post-mutation next-step hint to stderr.
func Next(msg string, args ...any) {
	if Quiet || HintsDisabled || os.Getenv("RC_NO_HINTS") != "" {
		return
	}
	if colorEnabled() {
		fmt.Fprintf(os.Stderr, colorYellow+"  next: "+msg+colorReset+"\n", args...)
	} else {
		fmt.Fprintf(os.Stderr, "  next: "+msg+"\n", args...)
	}
}

// Warn prints a warning message to stderr.
func Warn(msg string, args ...any) {
	if Quiet {
		return
	}
	if colorEnabled() {
		fmt.Fprintf(os.Stderr, colorYellow+"  Warning: "+msg+colorReset+"\n", args...)
	} else {
		fmt.Fprintf(os.Stderr, "  Warning: "+msg+"\n", args...)
	}
}

// Progress prints a progress indicator to stderr for bulk operations.
func Progress(current, total int, msg string, args ...any) {
	if Quiet {
		return
	}
	prefix := fmt.Sprintf("[%d/%d] ", current, total)
	fmt.Fprintf(os.Stderr, "  "+prefix+msg+"\n", args...)
}

// Log level constants.
const (
	LogLevelError = 0
	LogLevelWarn  = 1
	LogLevelInfo  = 2
	LogLevelDebug = 3
)

// LogLevel controls the verbosity of log output. Default is LogLevelWarn.
var LogLevel = LogLevelWarn

// Verbose is true when LogLevel >= LogLevelDebug. Kept for backward compatibility.
var Verbose bool

// ParseLogLevel converts a string to a log level constant.
func ParseLogLevel(s string) (int, bool) {
	switch strings.ToLower(s) {
	case "error":
		return LogLevelError, true
	case "warn":
		return LogLevelWarn, true
	case "info":
		return LogLevelInfo, true
	case "debug":
		return LogLevelDebug, true
	}
	return 0, false
}

// Info prints an informational message to stderr when log level >= info.
func Info(msg string, args ...any) {
	if LogLevel < LogLevelInfo || Quiet {
		return
	}
	fmt.Fprintf(os.Stderr, "  "+msg+"\n", args...)
}

// Debug prints a debug message to stderr when log level >= debug.
func Debug(msg string, args ...any) {
	if LogLevel < LogLevelDebug {
		return
	}
	if colorEnabled() {
		fmt.Fprintf(os.Stderr, "\033[90m  [debug] "+msg+colorReset+"\n", args...)
	} else {
		fmt.Fprintf(os.Stderr, "  [debug] "+msg+"\n", args...)
	}
}
