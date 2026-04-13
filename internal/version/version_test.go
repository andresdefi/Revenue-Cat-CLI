package version

import (
	"testing"
)

func TestVersionDefaults(t *testing.T) {
	// Default values are set at package level as var initializers.
	// In test/dev builds (no ldflags), they should have their defaults.
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if Commit == "" {
		t.Error("Commit should not be empty")
	}
	if Date == "" {
		t.Error("Date should not be empty")
	}
}

func TestVersion_DevDefault(t *testing.T) {
	// Without ldflags, Version defaults to "dev"
	if Version != "dev" {
		t.Logf("Version = %q (may be overridden by ldflags)", Version)
	}
}

func TestCommit_NoneDefault(t *testing.T) {
	// Without ldflags, Commit defaults to "none"
	if Commit != "none" {
		t.Logf("Commit = %q (may be overridden by ldflags)", Commit)
	}
}

func TestDate_UnknownDefault(t *testing.T) {
	// Without ldflags, Date defaults to "unknown"
	if Date != "unknown" {
		t.Logf("Date = %q (may be overridden by ldflags)", Date)
	}
}

func TestVersion_IsString(t *testing.T) {
	// Just verify they're accessible and non-panicking
	_ = Version
	_ = Commit
	_ = Date
}

func TestVersion_NotEmpty(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"Version", Version},
		{"Commit", Commit},
		{"Date", Date},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("%s should not be empty", tt.name)
			}
		})
	}
}
