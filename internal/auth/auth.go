package auth

import (
	"errors"
	"os"
	"sort"
	"strings"

	"github.com/andresdefi/rc/internal/cache"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/config"
	"github.com/zalando/go-keyring"
)

const (
	keychainService = "rc-cli"
)

var ErrNoToken = errors.New("not logged in - run `rc auth login` to authenticate")

// StoredProfile is a configured profile with a resolvable token.
type StoredProfile struct {
	Name   string
	Token  string
	Source string
}

// resolveProfile returns the effective profile name.
func resolveProfile(profile string) string {
	if profile != "" {
		return profile
	}
	return cmdutil.ResolveProfile()
}

// keychainUser returns the keychain account name for a profile.
func keychainUser(profile string) string {
	return "rc-cli:" + resolveProfile(profile)
}

func bypassKeychain() bool {
	return os.Getenv("RC_BYPASS_KEYCHAIN") == "1"
}

// SaveToken stores the API key in the system keychain, falling back to config file.
func SaveToken(profile, token string) error {
	profile = resolveProfile(profile)
	if !bypassKeychain() {
		err := keyring.Set(keychainService, keychainUser(profile), token)
		if err == nil {
			// Keep a config profile entry so keychain-backed profiles are discoverable.
			cfg, _ := config.Load()
			if cfg != nil {
				p := cfg.GetProfile(profile)
				if p == nil {
					p = &config.Profile{}
				}
				p.APIKey = ""
				cfg.SetProfile(profile, p)
				_ = config.Save(cfg)
			}
			_ = cache.Clear()
			return nil
		}
	}

	// Fallback to config file
	cfg, loadErr := config.Load()
	if loadErr != nil {
		return loadErr
	}
	p := cfg.GetProfile(profile)
	if p == nil {
		p = &config.Profile{}
	}
	p.APIKey = token
	cfg.SetProfile(profile, p)
	if err := config.Save(cfg); err != nil {
		return err
	}
	_ = cache.Clear()
	return nil
}

// GetToken retrieves the API key from keychain or config file for the given profile.
func GetToken(profile string) (string, error) {
	profile = resolveProfile(profile)
	if !bypassKeychain() {
		// Try keychain first
		token, err := keyring.Get(keychainService, keychainUser(profile))
		if err == nil && token != "" {
			return token, nil
		}
	}

	// Fallback to config file
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	p := cfg.GetProfile(profile)
	if p != nil && p.APIKey != "" {
		return p.APIKey, nil
	}

	return "", ErrNoToken
}

// DeleteToken removes the API key from both keychain and config file.
func DeleteToken(profile string) error {
	profile = resolveProfile(profile)
	if !bypassKeychain() {
		_ = keyring.Delete(keychainService, keychainUser(profile))
	}

	cfg, err := config.Load()
	if err != nil {
		return nil // If we can't load config, nothing to clear
	}
	p := cfg.GetProfile(profile)
	if p != nil && p.APIKey != "" {
		p.APIKey = ""
		cfg.SetProfile(profile, p)
		if err := config.Save(cfg); err != nil {
			return err
		}
	}
	_ = cache.Clear()
	return nil
}

// TokenSource returns where the token is stored ("keychain" or "config file").
func TokenSource(profile string) string {
	profile = resolveProfile(profile)
	if !bypassKeychain() {
		token, err := keyring.Get(keychainService, keychainUser(profile))
		if err == nil && token != "" {
			return "keychain"
		}
	}
	return "config file"
}

// StoredProfiles returns configured profiles that have a token in the keychain
// or config fallback. Profiles are sorted by name for stable output.
func StoredProfiles() ([]StoredProfile, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(cfg.Profiles)+1)
	names := make([]string, 0, len(cfg.Profiles)+1)
	addName := func(name string) {
		if name == "" {
			name = config.DefaultProfileName
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	for name := range cfg.Profiles {
		addName(name)
	}
	addName(cfg.CurrentProfile)
	sort.Strings(names)

	profiles := make([]StoredProfile, 0, len(names))
	for _, name := range names {
		token, err := GetToken(name)
		if err != nil {
			continue
		}
		profiles = append(profiles, StoredProfile{
			Name:   name,
			Token:  token,
			Source: TokenSource(name),
		})
	}
	return profiles, nil
}

// StoredProfileCount returns the number of configured profiles with tokens.
func StoredProfileCount() (int, error) {
	profiles, err := StoredProfiles()
	if err != nil {
		return 0, err
	}
	return len(profiles), nil
}

// MaskToken returns a masked version of the token for display.
func MaskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	prefix := token[:4]
	if strings.HasPrefix(token, "sk_") {
		prefix = token[:7] // Show "sk_XXXX"
	}
	return prefix + "..." + token[len(token)-4:]
}
