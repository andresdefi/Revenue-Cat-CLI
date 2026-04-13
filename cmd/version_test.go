package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCmd_Output(t *testing.T) {
	cmd := newVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Version writes to stdout directly via fmt.Printf, so we test via root
	root := NewRootCmd()
	root.SetArgs([]string{"version"})
	outBuf := new(bytes.Buffer)
	root.SetOut(outBuf)

	err := root.Execute()
	if err != nil {
		t.Fatalf("version command error: %v", err)
	}
}

func TestVersionCmd_ContainsVersionInfo(t *testing.T) {
	cmd := newVersionCmd()
	if cmd.Use != "version" {
		t.Errorf("Use = %q, want 'version'", cmd.Use)
	}
	if !strings.Contains(cmd.Short, "version") {
		t.Errorf("Short should contain 'version', got %q", cmd.Short)
	}
}
