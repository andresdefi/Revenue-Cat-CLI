package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/andresdefi/rc/internal/version"
)

func TestVersionCmd_Output(t *testing.T) {
	// Version writes to stdout directly via fmt.Printf, so we capture it
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	root := NewRootCmd()
	root.SetArgs([]string{"version"})

	err := root.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if err != nil {
		t.Fatalf("version command error: %v", err)
	}

	// Should contain "rc" prefix
	if !strings.HasPrefix(output, "rc ") {
		t.Errorf("version output should start with 'rc ', got %q", output)
	}

	// Should contain version value
	if !strings.Contains(output, version.Version) {
		t.Errorf("version output should contain version %q, got %q", version.Version, output)
	}

	// Should contain commit info
	if !strings.Contains(output, "commit:") {
		t.Errorf("version output should contain 'commit:', got %q", output)
	}

	// Should contain build date info
	if !strings.Contains(output, "built:") {
		t.Errorf("version output should contain 'built:', got %q", output)
	}
}

func TestVersionCmd_Use(t *testing.T) {
	cmd := newVersionCmd()
	if cmd.Use != "version" {
		t.Errorf("Use = %q, want 'version'", cmd.Use)
	}
}

func TestVersionCmd_Short(t *testing.T) {
	cmd := newVersionCmd()
	if cmd.Short == "" {
		t.Error("Short should not be empty")
	}
	if !strings.Contains(strings.ToLower(cmd.Short), "version") {
		t.Errorf("Short should mention 'version', got %q", cmd.Short)
	}
}

func TestVersionCmd_NotNil(t *testing.T) {
	cmd := newVersionCmd()
	if cmd == nil {
		t.Fatal("newVersionCmd() returned nil")
	}
}

func TestVersionCmd_HasRunFunc(t *testing.T) {
	cmd := newVersionCmd()
	if cmd.Run == nil && cmd.RunE == nil {
		t.Error("version command should have a Run or RunE function")
	}
}

func TestVersionCmd_NoArgs(t *testing.T) {
	cmd := newVersionCmd()
	// Version command takes no arguments - verify it runs with empty args
	if cmd.Args != nil {
		// If Args validator is set, it should accept 0 args
		if err := cmd.Args(cmd, []string{}); err != nil {
			t.Errorf("version should accept 0 args, got error: %v", err)
		}
	}
}

func TestVersionCmd_DefaultValues(t *testing.T) {
	// In test environment, version vars should have their defaults
	if version.Version == "" {
		t.Error("version.Version should not be empty")
	}
	if version.Commit == "" {
		t.Error("version.Commit should not be empty")
	}
	if version.Date == "" {
		t.Error("version.Date should not be empty")
	}
}

func TestVersionCmd_FormatPattern(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	root := NewRootCmd()
	root.SetArgs([]string{"version"})
	root.Execute()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	// Format should be: rc <version> (commit: <commit>, built: <date>)
	if !strings.Contains(output, "(") || !strings.Contains(output, ")") {
		t.Errorf("version output should contain parentheses, got %q", output)
	}
	if !strings.Contains(output, ", ") {
		t.Errorf("version output should contain comma separator, got %q", output)
	}
}

func TestVersionCmd_ExecuteDoesNotError(t *testing.T) {
	// Version command should always succeed, even without auth
	oldStdout := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	root := NewRootCmd()
	root.SetArgs([]string{"version"})
	err := root.Execute()

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("version command should not error, got: %v", err)
	}
}

func TestVersionCmd_FoundViaRoot(t *testing.T) {
	root := NewRootCmd()
	cmd, _, err := root.Find([]string{"version"})
	if err != nil {
		t.Fatalf("Find version: %v", err)
	}
	if cmd.Name() != "version" {
		t.Errorf("found command name = %q, want %q", cmd.Name(), "version")
	}
}
