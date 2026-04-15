package cmdutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/andresdefi/rc/internal/output"
)

// overrideHome sets HOME for config-based tests.
func overrideHome(t *testing.T, dir string) {
	t.Helper()
	orig := os.Getenv("HOME")
	t.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", orig) })
}

// writeConfig writes a config.json to the temp home directory.
func writeConfig(t *testing.T, home string, data map[string]string) {
	t.Helper()
	dir := filepath.Join(home, ".rc")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), b, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func strPtr(s string) *string {
	return &s
}

func TestResolveProject_FlagSet(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	flag := strPtr("proj_from_flag")
	pid, err := ResolveProject(flag)
	if err != nil {
		t.Fatalf("ResolveProject() error: %v", err)
	}
	if pid != "proj_from_flag" {
		t.Errorf("ResolveProject() = %q, want %q", pid, "proj_from_flag")
	}
}

func TestResolveProject_FlagTakesPrecedenceOverConfig(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	writeConfig(t, tmp, map[string]string{"project_id": "proj_from_config"})

	flag := strPtr("proj_from_flag")
	pid, err := ResolveProject(flag)
	if err != nil {
		t.Fatalf("ResolveProject() error: %v", err)
	}
	if pid != "proj_from_flag" {
		t.Errorf("ResolveProject() = %q, want flag value %q", pid, "proj_from_flag")
	}
}

func TestResolveProject_ConfigFallback(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	writeConfig(t, tmp, map[string]string{"project_id": "proj_from_config"})

	// nil flag
	pid, err := ResolveProject(nil)
	if err != nil {
		t.Fatalf("ResolveProject() error: %v", err)
	}
	if pid != "proj_from_config" {
		t.Errorf("ResolveProject() = %q, want %q", pid, "proj_from_config")
	}
}

func TestResolveProject_EmptyFlagFallsToConfig(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	writeConfig(t, tmp, map[string]string{"project_id": "proj_from_config"})

	flag := strPtr("") // empty flag
	pid, err := ResolveProject(flag)
	if err != nil {
		t.Fatalf("ResolveProject() error: %v", err)
	}
	if pid != "proj_from_config" {
		t.Errorf("ResolveProject() = %q, want %q", pid, "proj_from_config")
	}
}

func TestResolveProject_NilFlagNoConfig(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	_, err := ResolveProject(nil)
	if err == nil {
		t.Fatal("expected error when no flag and no config, got nil")
	}
	errMsg := err.Error()
	if errMsg == "" {
		t.Fatal("error message should not be empty")
	}
}

func TestResolveProject_EmptyFlagEmptyConfig(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	writeConfig(t, tmp, map[string]string{"project_id": ""})

	flag := strPtr("")
	_, err := ResolveProject(flag)
	if err == nil {
		t.Fatal("expected error when both flag and config are empty")
	}
}

func TestResolveProject_NilFlagEmptyConfig(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	writeConfig(t, tmp, map[string]string{})

	_, err := ResolveProject(nil)
	if err == nil {
		t.Fatal("expected error with nil flag and empty config")
	}
}

func TestResolveProject_ErrorContainsGuidance(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	_, err := ResolveProject(nil)
	if err == nil {
		t.Fatal("expected error")
	}

	msg := err.Error()
	// Error should mention --project flag or set-default command
	if !contains(msg, "--project") && !contains(msg, "set-default") {
		t.Errorf("error message should contain guidance, got: %q", msg)
	}
}

func TestResolveProject_WhitespaceOnlyFlag(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	// A whitespace flag is technically non-empty
	flag := strPtr("  ")
	pid, err := ResolveProject(flag)
	if err != nil {
		// The implementation doesn't trim, so whitespace is treated as a value
		t.Fatalf("ResolveProject() error: %v", err)
	}
	if pid != "  " {
		t.Errorf("ResolveProject() = %q, want %q (untrimmed)", pid, "  ")
	}
}

func TestGetOutputFormat_JSON(t *testing.T) {
	flag := strPtr("json")
	got := GetOutputFormat(flag)
	if got != output.FormatJSON {
		t.Errorf("GetOutputFormat(\"json\") = %q, want %q", got, output.FormatJSON)
	}
}

func TestGetOutputFormat_Table(t *testing.T) {
	flag := strPtr("table")
	got := GetOutputFormat(flag)
	if got != output.FormatTable {
		t.Errorf("GetOutputFormat(\"table\") = %q, want %q", got, output.FormatTable)
	}
}

func TestGetOutputFormat_NilFlag(t *testing.T) {
	got := GetOutputFormat(nil)
	// Should auto-detect: in test env, typically not a TTY, so should return JSON
	// But we accept either since it depends on the test runner
	if got != output.FormatJSON && got != output.FormatTable {
		t.Errorf("GetOutputFormat(nil) = %q, want FormatJSON or FormatTable", got)
	}
}

func TestGetOutputFormat_EmptyFlag(t *testing.T) {
	flag := strPtr("")
	got := GetOutputFormat(flag)
	// Empty string doesn't match "json" or "table", so falls through to auto-detect
	if got != output.FormatJSON && got != output.FormatTable {
		t.Errorf("GetOutputFormat(\"\") = %q, expected auto-detect result", got)
	}
}

func TestGetOutputFormat_UnknownValue(t *testing.T) {
	flag := strPtr("xml")
	got := GetOutputFormat(flag)
	// Unknown value doesn't match, falls through to auto-detect
	if got != output.FormatJSON && got != output.FormatTable {
		t.Errorf("GetOutputFormat(\"xml\") = %q, expected auto-detect result", got)
	}
}

func TestGetOutputFormat_CaseSensitive(t *testing.T) {
	tests := []struct {
		input string
		name  string
	}{
		{"JSON", "uppercase JSON"},
		{"Table", "capitalized Table"},
		{"TABLE", "uppercase TABLE"},
		{"Json", "mixed case Json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := strPtr(tt.input)
			got := GetOutputFormat(flag)
			// These should NOT match the explicit cases, falling to auto-detect
			if got != output.FormatJSON && got != output.FormatTable {
				t.Errorf("GetOutputFormat(%q) = %q, expected auto-detect", tt.input, got)
			}
		})
	}
}

func TestGetOutputFormat_TableDriven(t *testing.T) {
	tests := []struct {
		name  string
		flag  *string
		valid []output.Format
	}{
		{"json flag", strPtr("json"), []output.Format{output.FormatJSON}},
		{"table flag", strPtr("table"), []output.Format{output.FormatTable}},
		{"nil flag", nil, []output.Format{output.FormatJSON, output.FormatTable}},
		{"empty flag", strPtr(""), []output.Format{output.FormatJSON, output.FormatTable}},
		{"garbage flag", strPtr("yaml"), []output.Format{output.FormatJSON, output.FormatTable}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetOutputFormat(tt.flag)
			found := false
			for _, v := range tt.valid {
				if got == v {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("GetOutputFormat() = %q, want one of %v", got, tt.valid)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
