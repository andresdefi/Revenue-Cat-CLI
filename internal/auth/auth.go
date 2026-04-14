package auth

import (
	"errors"
	"strings"

	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/config"
	"github.com/zalando/go-keyring"
)

const (
	keychainService = "rc-cli"
)

var ErrNoToken = errors.New("not logged in - run `rc auth login` to authenticate")

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

// SaveToken stores the API key in the system keychain, falling back to config file.
func SaveToken(profile, token string) error {
	profile = resolveProfile(profile)
	err := keyring.Set(keychainService, keychainUser(profile), token)
	if err == nil {
		// Clear any config file token since keychain is preferred
		cfg, _ := config.Load()
		if cfg != nil {
			p := cfg.GetProfile(profile)
			if p != nil && p.APIKey != "" {
				p.APIKey = ""
				cfg.SetProfile(profile, p)
				_ = config.Save(cfg)
			}
		}
		return nil
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
	return config.Save(cfg)
}

// GetToken retrieves the API key from keychain or config file for the given profile.
func GetToken(profile string) (string, error) {
	profile = resolveProfile(profile)
	// Try keychain first
	token, err := keyring.Get(keychainService, keychainUser(profile))
	if err == nil && token != "" {
		return token, nil
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
	_ = keyring.Delete(keychainService, keychainUser(profile))

	cfg, err := config.Load()
	if err != nil {
		return nil // If we can't load config, nothing to clear
	}
	p := cfg.GetProfile(profile)
	if p != nil && p.APIKey != "" {
		p.APIKey = ""
		cfg.SetProfile(profile, p)
		return config.Save(cfg)
	}
	return nil
}

// TokenSource returns where the token is stored ("keychain" or "config file").
func TokenSource(profile string) string {
	profile = resolveProfile(profile)
	token, err := keyring.Get(keychainService, keychainUser(profile))
	if err == nil && token != "" {
		return "keychain"
	}
	return "config file"
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
