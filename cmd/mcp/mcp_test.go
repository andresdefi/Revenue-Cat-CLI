package mcp

import (
	"testing"
)

func TestNewMCPCmd_NotNil(t *testing.T) {
	cmd := NewMCPCmd()
	if cmd == nil {
		t.Fatal("NewMCPCmd() returned nil")
	}
}

func TestNewMCPCmd_Use(t *testing.T) {
	cmd := NewMCPCmd()
	if cmd.Use != "mcp" {
		t.Errorf("Use = %q, want %q", cmd.Use, "mcp")
	}
}

func TestNewMCPCmd_Short(t *testing.T) {
	cmd := NewMCPCmd()
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestNewMCPCmd_HasServeSubcommand(t *testing.T) {
	cmd := NewMCPCmd()

	subNames := make(map[string]bool)
	for _, c := range cmd.Commands() {
		subNames[c.Name()] = true
	}

	if !subNames["serve"] {
		t.Error("mcp command should have 'serve' subcommand")
	}
}

func TestNewMCPCmd_ServeSubcommandNotNil(t *testing.T) {
	cmd := NewMCPCmd()

	for _, c := range cmd.Commands() {
		if c.Name() == "serve" {
			if c.Short == "" {
				t.Error("serve subcommand Short should not be empty")
			}
			if c.Long == "" {
				t.Error("serve subcommand Long should not be empty")
			}
			return
		}
	}
	t.Fatal("serve subcommand not found")
}

func TestNewMCPCmd_SubcommandCount(t *testing.T) {
	cmd := NewMCPCmd()
	commands := cmd.Commands()
	if len(commands) < 1 {
		t.Errorf("mcp command has %d subcommands, want >= 1", len(commands))
	}
}
