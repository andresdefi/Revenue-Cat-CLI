package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// overrideHome temporarily sets HOME to dir for the duration of the test.
func overrideHome(t *testing.T, dir string) {
	t.Helper()
	orig := os.Getenv("HOME")
	t.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", orig) })
}

func TestDir(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	dir, err := Dir()
	if err != nil {
		t.Fatalf("Dir() error: %v", err)
	}
	want := filepath.Join(tmp, ".rc")
	if dir != want {
		t.Errorf("Dir() = %q, want %q", dir, want)
	}
}

func TestPath(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	p, err := Path()
	if err != nil {
		t.Fatalf("Path() error: %v", err)
	}
	want := filepath.Join(tmp, ".rc", "config.json")
	if p != want {
		t.Errorf("Path() = %q, want %q", p, want)
	}
}

func TestLoad_NoFile(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if cfg.APIKey != "" {
		t.Errorf("expected empty APIKey, got %q", cfg.APIKey)
	}
	if cfg.ProjectID != "" {
		t.Errorf("expected empty ProjectID, got %q", cfg.ProjectID)
	}
}

func TestLoad_ExistingFile(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	dir := filepath.Join(tmp, ".rc")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	data := `{"api_key":"sk_test_abc123","project_id":"proj_xyz"}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(data), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.APIKey != "sk_test_abc123" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "sk_test_abc123")
	}
	if cfg.ProjectID != "proj_xyz" {
		t.Errorf("ProjectID = %q, want %q", cfg.ProjectID, "proj_xyz")
	}
}

func TestLoad_MalformedJSON(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	dir := filepath.Join(tmp, ".rc")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{invalid json"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	dir := filepath.Join(tmp, ".rc")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(""), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
}

func TestLoad_EmptyObject(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	dir := filepath.Join(tmp, ".rc")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{}"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.APIKey != "" {
		t.Errorf("expected empty APIKey, got %q", cfg.APIKey)
	}
	if cfg.ProjectID != "" {
		t.Errorf("expected empty ProjectID, got %q", cfg.ProjectID)
	}
}

func TestLoad_ExtraFields(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	dir := filepath.Join(tmp, ".rc")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	data := `{"api_key":"sk_test","project_id":"proj1","extra_field":"ignored"}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(data), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.APIKey != "sk_test" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "sk_test")
	}
}

func TestLoad_OnlyAPIKey(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	dir := filepath.Join(tmp, ".rc")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	data := `{"api_key":"sk_key_only"}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(data), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.APIKey != "sk_key_only" {
		t.Errorf("APIKey = %q, want %q", cfg.APIKey, "sk_key_only")
	}
	if cfg.ProjectID != "" {
		t.Errorf("expected empty ProjectID, got %q", cfg.ProjectID)
	}
}

func TestSave_CreatesDir(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	cfg := &Config{
		APIKey:    "sk_test_save",
		ProjectID: "proj_save",
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Verify directory was created
	dir := filepath.Join(tmp, ".rc")
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("config dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("config dir is not a directory")
	}

	// Verify file contents
	path := filepath.Join(dir, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if loaded.APIKey != "sk_test_save" {
		t.Errorf("APIKey = %q, want %q", loaded.APIKey, "sk_test_save")
	}
	if loaded.ProjectID != "proj_save" {
		t.Errorf("ProjectID = %q, want %q", loaded.ProjectID, "proj_save")
	}
}

func TestSave_OverwritesExisting(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	// Save initial config
	cfg1 := &Config{APIKey: "key1", ProjectID: "proj1"}
	if err := Save(cfg1); err != nil {
		t.Fatalf("Save(cfg1) error: %v", err)
	}

	// Save updated config
	cfg2 := &Config{APIKey: "key2", ProjectID: "proj2"}
	if err := Save(cfg2); err != nil {
		t.Fatalf("Save(cfg2) error: %v", err)
	}

	// Verify overwritten
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if loaded.APIKey != "key2" {
		t.Errorf("APIKey = %q, want %q", loaded.APIKey, "key2")
	}
	if loaded.ProjectID != "proj2" {
		t.Errorf("ProjectID = %q, want %q", loaded.ProjectID, "proj2")
	}
}

func TestSave_OmitsEmptyFields(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	cfg := &Config{} // all empty
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	path := filepath.Join(tmp, ".rc", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// With omitempty, the JSON should be essentially empty object
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(raw) != 0 {
		t.Errorf("expected empty JSON object, got %v", raw)
	}
}

func TestSave_FilePermissions(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	cfg := &Config{APIKey: "sk_secret"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	path := filepath.Join(tmp, ".rc", "config.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestSave_DirPermissions(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	cfg := &Config{APIKey: "sk_secret"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	dir := filepath.Join(tmp, ".rc")
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0700 {
		t.Errorf("dir permissions = %o, want 0700", perm)
	}
}

func TestSave_IndentedJSON(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	cfg := &Config{APIKey: "sk_test", ProjectID: "proj1"}
	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	path := filepath.Join(tmp, ".rc", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	expected, _ := json.MarshalIndent(cfg, "", "  ")
	if string(data) != string(expected) {
		t.Errorf("file content mismatch.\ngot:\n%s\nwant:\n%s", data, expected)
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	original := &Config{
		APIKey:    "sk_roundtrip_key_123456789",
		ProjectID: "projABCDEF",
	}

	if err := Save(original); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.APIKey != original.APIKey {
		t.Errorf("APIKey = %q, want %q", loaded.APIKey, original.APIKey)
	}
	if loaded.ProjectID != original.ProjectID {
		t.Errorf("ProjectID = %q, want %q", loaded.ProjectID, original.ProjectID)
	}
}

func TestConfig_JSONSerialization(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want map[string]string
	}{
		{
			name: "full config",
			cfg:  Config{APIKey: "sk_abc", ProjectID: "proj_123"},
			want: map[string]string{"api_key": "sk_abc", "project_id": "proj_123"},
		},
		{
			name: "only api key",
			cfg:  Config{APIKey: "sk_abc"},
			want: map[string]string{"api_key": "sk_abc"},
		},
		{
			name: "only project id",
			cfg:  Config{ProjectID: "proj_123"},
			want: map[string]string{"project_id": "proj_123"},
		},
		{
			name: "empty config",
			cfg:  Config{},
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.cfg)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}

			var raw map[string]string
			if err := json.Unmarshal(data, &raw); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}

			if len(raw) != len(tt.want) {
				t.Errorf("field count = %d, want %d; got %v", len(raw), len(tt.want), raw)
			}
			for k, v := range tt.want {
				if raw[k] != v {
					t.Errorf("field %q = %q, want %q", k, raw[k], v)
				}
			}
		})
	}
}
