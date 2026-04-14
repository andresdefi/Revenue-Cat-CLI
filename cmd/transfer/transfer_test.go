package transfer

import (
	"testing"
)

func TestNewExportCmd_NotNil(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := NewExportCmd(&projectID, &outputFormat)
	if cmd == nil {
		t.Fatal("NewExportCmd() returned nil")
	}
}

func TestNewExportCmd_Use(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := NewExportCmd(&projectID, &outputFormat)
	if cmd.Use != "export" {
		t.Errorf("Use = %q, want %q", cmd.Use, "export")
	}
}

func TestNewExportCmd_Short(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := NewExportCmd(&projectID, &outputFormat)
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestNewExportCmd_HasFileFlag(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := NewExportCmd(&projectID, &outputFormat)
	f := cmd.Flags().Lookup("file")
	if f == nil {
		t.Fatal("export command missing --file flag")
	}
	if f.DefValue != "" {
		t.Errorf("--file default = %q, want empty string", f.DefValue)
	}
}

func TestNewImportCmd_NotNil(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := NewImportCmd(&projectID, &outputFormat)
	if cmd == nil {
		t.Fatal("NewImportCmd() returned nil")
	}
}

func TestNewImportCmd_Use(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := NewImportCmd(&projectID, &outputFormat)
	if cmd.Use != "import" {
		t.Errorf("Use = %q, want %q", cmd.Use, "import")
	}
}

func TestNewImportCmd_Short(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := NewImportCmd(&projectID, &outputFormat)
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestNewImportCmd_HasFileFlag(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := NewImportCmd(&projectID, &outputFormat)
	f := cmd.Flags().Lookup("file")
	if f == nil {
		t.Fatal("import command missing --file flag")
	}
	if f.DefValue != "" {
		t.Errorf("--file default = %q, want empty string", f.DefValue)
	}
}

func TestNewExportCmd_Long(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := NewExportCmd(&projectID, &outputFormat)
	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestNewImportCmd_Long(t *testing.T) {
	projectID := ""
	outputFormat := ""
	cmd := NewImportCmd(&projectID, &outputFormat)
	if cmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}
