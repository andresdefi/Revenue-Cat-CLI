package auth

import (
	"errors"
	"strings"

	"github.com/andresdefi/rc/internal/config"
	"github.com/zalando/go-keyring"
)

const (
	keychainService = "rc-cli"
	keychainUser    = "api-key"
)

var ErrNoToken = errors.New("not logged in - run `rc auth login` to authenticate")

// SaveToken stores the API key in the system keychain, falling back to config file.
func SaveToken(token string) error {
	err := keyring.Set(keychainService, keychainUser, token)
	if err == nil {
		// Clear any config file token since keychain is preferred
		cfg, _ := config.Load()
		if cfg != nil && cfg.APIKey != "" {
			cfg.APIKey = ""
			_ = config.Save(cfg)
		}
		return nil
	}

	// Fallback to config file
	cfg, loadErr := config.Load()
	if loadErr != nil {
		return loadErr
	}
	cfg.APIKey = token
	return config.Save(cfg)
}

// GetToken retrieves the API key from keychain or config file.
func GetToken() (string, error) {
	// Try keychain first
	token, err := keyring.Get(keychainService, keychainUser)
	if err == nil && token != "" {
		return token, nil
	}

	// Fallback to config file
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}
	if cfg.APIKey != "" {
		return cfg.APIKey, nil
	}

	return "", ErrNoToken
}

// DeleteToken removes the API key from both keychain and config file.
func DeleteToken() error {
	_ = keyring.Delete(keychainService, keychainUser)

	cfg, err := config.Load()
	if err != nil {
		return nil // If we can't load config, nothing to clear
	}
	if cfg.APIKey != "" {
		cfg.APIKey = ""
		return config.Save(cfg)
	}
	return nil
}

// TokenSource returns where the token is stored ("keychain" or "config file").
func TokenSource() string {
	token, err := keyring.Get(keychainService, keychainUser)
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
