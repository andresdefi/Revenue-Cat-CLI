package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestDir(t *testing.T) {
	home := setupTestHome(t)
	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}
	if dir != filepath.Join(home, ".rc") {
		t.Errorf("Dir() = %q, want %q", dir, filepath.Join(home, ".rc"))
	}
}

func TestPath(t *testing.T) {
	home := setupTestHome(t)
	path, err := Path()
	if err != nil {
		t.Fatalf("Path() error: %v", err)
	}
	expected := filepath.Join(home, ".rc", "config.toml")
	if path != expected {
		t.Errorf("Path() = %q, want %q", path, expected)
	}
}

func TestLoad_NoFile(t *testing.T) {
	setupTestHome(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.CurrentProfile != DefaultProfileName {
		t.Errorf("CurrentProfile = %q, want %q", cfg.CurrentProfile, DefaultProfileName)
	}
	if cfg.Profiles == nil {
		t.Fatal("Profiles should not be nil")
	}
}

func TestSave_And_Load(t *testing.T) {
	setupTestHome(t)
	cfg := &Config{
		CurrentProfile: "prod",
		Profiles: map[string]*Profile{
			"prod": {
				APIKey:    "sk_test_123",
				ProjectID: "proj_abc",
			},
			"staging": {
				ProjectID: "proj_staging",
			},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.CurrentProfile != "prod" {
		t.Errorf("CurrentProfile = %q, want %q", loaded.CurrentProfile, "prod")
	}
	if loaded.Profiles["prod"].APIKey != "sk_test_123" {
		t.Errorf("prod API key = %q, want %q", loaded.Profiles["prod"].APIKey, "sk_test_123")
	}
	if loaded.Profiles["prod"].ProjectID != "proj_abc" {
		t.Errorf("prod ProjectID = %q, want %q", loaded.Profiles["prod"].ProjectID, "proj_abc")
	}
	if loaded.Profiles["staging"].ProjectID != "proj_staging" {
		t.Errorf("staging ProjectID = %q, want %q", loaded.Profiles["staging"].ProjectID, "proj_staging")
	}
}

func TestActiveProfile(t *testing.T) {
	cfg := &Config{
		CurrentProfile: "work",
		Profiles: map[string]*Profile{
			"work": {ProjectID: "proj_work"},
		},
	}
	p := cfg.ActiveProfile()
	if p.ProjectID != "proj_work" {
		t.Errorf("ActiveProfile().ProjectID = %q, want %q", p.ProjectID, "proj_work")
	}
}

func TestActiveProfile_FallsBackToDefault(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]*Profile{
			DefaultProfileName: {ProjectID: "proj_default"},
		},
	}
	p := cfg.ActiveProfile()
	if p.ProjectID != "proj_default" {
		t.Errorf("ActiveProfile().ProjectID = %q, want %q", p.ProjectID, "proj_default")
	}
}

func TestActiveProfile_EmptyWhenMissing(t *testing.T) {
	cfg := &Config{}
	p := cfg.ActiveProfile()
	if p.ProjectID != "" {
		t.Errorf("ActiveProfile().ProjectID = %q, want empty", p.ProjectID)
	}
}

func TestGetProfile(t *testing.T) {
	cfg := &Config{
		Profiles: map[string]*Profile{
			"test": {APIKey: "key123"},
		},
	}
	p := cfg.GetProfile("test")
	if p == nil || p.APIKey != "key123" {
		t.Error("GetProfile should return the profile")
	}
	if cfg.GetProfile("nonexistent") != nil {
		t.Error("GetProfile should return nil for missing profile")
	}
}

func TestSetProfile(t *testing.T) {
	cfg := &Config{}
	cfg.SetProfile("new", &Profile{ProjectID: "proj_new"})
	if cfg.Profiles["new"].ProjectID != "proj_new" {
		t.Error("SetProfile should create the profile")
	}
}

func TestLegacyMigration(t *testing.T) {
	home := setupTestHome(t)
	rcDir := filepath.Join(home, ".rc")
	os.MkdirAll(rcDir, 0700)

	legacy := `{"api_key": "sk_old_key", "project_id": "proj_old"}`
	os.WriteFile(filepath.Join(rcDir, "config.json"), []byte(legacy), 0600)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	p := cfg.GetProfile(DefaultProfileName)
	if p == nil {
		t.Fatal("default profile should exist after migration")
	}
	if p.APIKey != "sk_old_key" {
		t.Errorf("migrated API key = %q, want %q", p.APIKey, "sk_old_key")
	}
	if p.ProjectID != "proj_old" {
		t.Errorf("migrated ProjectID = %q, want %q", p.ProjectID, "proj_old")
	}

	if _, err := os.Stat(filepath.Join(rcDir, "config.json")); !os.IsNotExist(err) {
		t.Error("legacy config.json should be removed after migration")
	}
	if _, err := os.Stat(filepath.Join(rcDir, "config.toml")); err != nil {
		t.Error("config.toml should exist after migration")
	}
}

func TestSave_FilePermissions(t *testing.T) {
	setupTestHome(t)
	cfg := &Config{
		CurrentProfile: DefaultProfileName,
		Profiles:       map[string]*Profile{DefaultProfileName: {}},
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	path, _ := Path()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat error: %v", err)
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("config file permissions = %o, want 0600", perm)
	}
}
