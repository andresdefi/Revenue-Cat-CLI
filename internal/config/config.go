package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	configDirName      = ".rc"
	configFileName     = "config.toml"
	legacyConfigName   = "config.json"
	DefaultProfileName = "default"
)

// Config is the top-level TOML config with multi-profile support.
type Config struct {
	CurrentProfile string              `toml:"current_profile"`
	Profiles       map[string]*Profile `toml:"profiles"`
}

// Profile holds per-profile settings.
type Profile struct {
	APIKey    string `toml:"api_key,omitempty"`
	ProjectID string `toml:"project_id,omitempty"`
}

// ActiveProfile returns the profile identified by CurrentProfile,
// or an empty profile if none is set.
func (c *Config) ActiveProfile() *Profile {
	if c.Profiles == nil {
		return &Profile{}
	}
	name := c.CurrentProfile
	if name == "" {
		name = DefaultProfileName
	}
	p, ok := c.Profiles[name]
	if !ok {
		return &Profile{}
	}
	return p
}

// GetProfile returns the named profile or nil.
func (c *Config) GetProfile(name string) *Profile {
	if c.Profiles == nil {
		return nil
	}
	return c.Profiles[name]
}

// SetProfile upserts a profile by name.
func (c *Config) SetProfile(name string, p *Profile) {
	if c.Profiles == nil {
		c.Profiles = make(map[string]*Profile)
	}
	c.Profiles[name] = p
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDirName), nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// legacyPath returns the path to the old config.json.
func legacyPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, legacyConfigName), nil
}

// legacyConfig is the old JSON config shape used for migration.
type legacyConfig struct {
	APIKey    string `json:"api_key,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
}

// migrateIfNeeded checks for a legacy config.json and converts it to config.toml.
func migrateIfNeeded() (*Config, bool, error) {
	lp, err := legacyPath()
	if err != nil {
		return nil, false, err
	}
	tp, err := Path()
	if err != nil {
		return nil, false, err
	}

	// Only migrate when legacy exists and TOML does not.
	if _, err := os.Stat(tp); err == nil {
		return nil, false, nil // TOML already exists
	}
	data, err := os.ReadFile(lp)
	if err != nil {
		return nil, false, nil // no legacy file
	}

	var old legacyConfig
	if err := json.Unmarshal(data, &old); err != nil {
		return nil, false, nil // malformed legacy, ignore
	}

	cfg := &Config{
		CurrentProfile: DefaultProfileName,
		Profiles: map[string]*Profile{
			DefaultProfileName: {
				APIKey:    old.APIKey,
				ProjectID: old.ProjectID,
			},
		},
	}

	if err := Save(cfg); err != nil {
		return nil, false, err
	}

	// Remove legacy file after successful migration.
	_ = os.Remove(lp)
	return cfg, true, nil
}

func Load() (*Config, error) {
	// Try migration first.
	migrated, ok, err := migrateIfNeeded()
	if err != nil {
		return nil, err
	}
	if ok {
		return migrated, nil
	}

	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				CurrentProfile: DefaultProfileName,
				Profiles:       map[string]*Profile{DefaultProfileName: {}},
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]*Profile)
	}
	if cfg.CurrentProfile == "" {
		cfg.CurrentProfile = DefaultProfileName
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	path, err := Path()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	enc := toml.NewEncoder(f)
	if err := enc.Encode(cfg); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return closeErr
		}
		return err
	}
	return f.Close()
}
