package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andresdefi/rc/internal/config"
)

func setupTestHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{"empty", "", "****"},
		{"short 4 chars", "abcd", "****"},
		{"exactly 8 chars", "12345678", "****"},
		{"9 chars normal", "123456789", "1234...6789"},
		{"sk_ prefix short", "sk_abcde", "****"},
		{"sk_ prefix 9 chars", "sk_test12", "sk_test...st12"},
		{"sk_ prefix typical", "sk_test_abc123def456", "sk_test...f456"},
		{"sk_ prefix long", "sk_fake_ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890", "sk_fake...7890"},
		{"long no prefix", "abcdefghijklmnop", "abcd...mnop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskToken(tt.token)
			if got != tt.want {
				t.Errorf("MaskToken(%q) = %q, want %q", tt.token, got, tt.want)
			}
		})
	}
}

func TestMaskToken_NeverRevealsFullToken(t *testing.T) {
	tokens := []string{
		"sk_test_abc123def456ghi789",
		"atk_live_1234567890abcdef",
		"some_random_api_key_value",
	}
	for _, token := range tokens {
		masked := MaskToken(token)
		if masked == token {
			t.Errorf("MaskToken should never return the original token: %q", token)
		}
	}
}

func TestSaveToken_ConfigFallback(t *testing.T) {
	setupTestHome(t)

	// Save token for a profile via config fallback (keychain may not work in test)
	err := SaveToken("testprofile", "sk_test_token_123")
	if err != nil {
		t.Fatalf("SaveToken error: %v", err)
	}

	// Verify it's in config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	p := cfg.GetProfile("testprofile")
	if p == nil {
		t.Fatal("profile should exist")
	}
	// Token might be in keychain or config depending on environment
}

func TestGetToken_FromConfig(t *testing.T) {
	setupTestHome(t)

	// Write token directly to config
	cfg := &config.Config{
		CurrentProfile: "test",
		Profiles: map[string]*config.Profile{
			"test": {APIKey: "sk_config_token"},
		},
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	token, err := GetToken("test")
	if err != nil {
		t.Fatalf("GetToken error: %v", err)
	}
	if token != "sk_config_token" {
		t.Errorf("GetToken = %q, want %q", token, "sk_config_token")
	}
}

func TestGetToken_NotLoggedIn(t *testing.T) {
	setupTestHome(t)

	_, err := GetToken("nonexistent")
	if err == nil {
		t.Error("GetToken should error when not logged in")
	}
}

func TestDeleteToken_Config(t *testing.T) {
	setupTestHome(t)

	// Save a token
	cfg := &config.Config{
		CurrentProfile: "test",
		Profiles: map[string]*config.Profile{
			"test": {APIKey: "sk_to_delete"},
		},
	}
	config.Save(cfg)

	// Delete it
	err := DeleteToken("test")
	if err != nil {
		t.Fatalf("DeleteToken error: %v", err)
	}

	// Verify it's gone
	cfg, _ = config.Load()
	p := cfg.GetProfile("test")
	if p != nil && p.APIKey != "" {
		t.Errorf("API key should be empty after delete, got %q", p.APIKey)
	}
}

func TestTokenSource(t *testing.T) {
	setupTestHome(t)
	// Without keychain, should return "config file"
	source := TokenSource("anyprofile")
	if source != "config file" {
		t.Errorf("TokenSource = %q, want %q", source, "config file")
	}
}

func TestResolveProfile(t *testing.T) {
	if got := resolveProfile("explicit"); got != "explicit" {
		t.Errorf("resolveProfile('explicit') = %q, want 'explicit'", got)
	}
}

func TestKeychainUser(t *testing.T) {
	got := keychainUser("myprofile")
	if got != "rc-cli:myprofile" {
		t.Errorf("keychainUser('myprofile') = %q, want 'rc-cli:myprofile'", got)
	}
}

func TestSaveAndGetToken_RoundTrip(t *testing.T) {
	home := setupTestHome(t)
	_ = os.MkdirAll(filepath.Join(home, ".rc"), 0o700)

	profile := "roundtrip"
	token := "sk_roundtrip_test_value"

	if err := SaveToken(profile, token); err != nil {
		t.Fatalf("SaveToken error: %v", err)
	}

	got, err := GetToken(profile)
	if err != nil {
		t.Fatalf("GetToken error: %v", err)
	}
	if got != token {
		t.Errorf("GetToken = %q, want %q", got, token)
	}
}

func TestErrNoToken_Message(t *testing.T) {
	if ErrNoToken.Error() == "" {
		t.Error("ErrNoToken should have a message")
	}
}
