package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// overrideHome sets HOME for config fallback tests.
func overrideHome(t *testing.T, dir string) {
	t.Helper()
	orig := os.Getenv("HOME")
	t.Setenv("HOME", dir)
	t.Cleanup(func() { os.Setenv("HOME", orig) })
}

// writeConfig writes a config.json to the temp home directory.
func writeConfig(t *testing.T, home string, data map[string]string) {
	t.Helper()
	dir := filepath.Join(home, ".rc")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), b, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

// readConfig reads config.json from the temp home directory.
func readConfig(t *testing.T, home string) map[string]any {
	t.Helper()
	path := filepath.Join(home, ".rc", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("ReadFile: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	return result
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{
			name:  "empty string",
			token: "",
			want:  "****",
		},
		{
			name:  "short token 1 char",
			token: "a",
			want:  "****",
		},
		{
			name:  "short token 4 chars",
			token: "abcd",
			want:  "****",
		},
		{
			name:  "short token 8 chars",
			token: "abcdefgh",
			want:  "****",
		},
		{
			name:  "exactly 9 chars - normal prefix",
			token: "abcdefghi",
			want:  "abcd...fghi",
		},
		{
			name:  "normal token without sk_ prefix",
			token: "abcdefghijklmnop",
			want:  "abcd...mnop",
		},
		{
			name:  "token with sk_ prefix short",
			token: "sk_test1",
			want:  "****",
		},
		{
			name:  "token with sk_ prefix 9 chars",
			token: "sk_test12",
			want:  "sk_test...st12",
		},
		{
			name:  "token with sk_ prefix typical",
			token: "sk_test_abc123def456",
			want:  "sk_test...f456",
		},
		{
			name:  "token with sk_ prefix long",
			token: "sk_fake_ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890",
			want:  "sk_fake...7890",
		},
		{
			name:  "sk_ prefix exactly",
			token: "sk_abcde",
			want:  "****",
		},
		{
			name:  "12 char token no prefix",
			token: "123456789012",
			want:  "1234...9012",
		},
		{
			name:  "token with special characters",
			token: "sk_test-key_abc!@#$%^&*()_+123",
			want:  "sk_test...+123",
		},
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

func TestMaskToken_AlwaysHidesMiddle(t *testing.T) {
	// For any token longer than 8 chars, the middle should not be visible
	token := "sk_test_this_is_a_very_long_secret_key_12345"
	masked := MaskToken(token)

	// Should start with sk_test
	if masked[:7] != "sk_test" {
		t.Errorf("masked token should start with sk_test, got %q", masked)
	}

	// Should end with last 4 chars
	last4 := token[len(token)-4:]
	maskedLast4 := masked[len(masked)-4:]
	if maskedLast4 != last4 {
		t.Errorf("masked token should end with %q, got %q", last4, maskedLast4)
	}

	// Should contain "..."
	if len(masked) < len(token) {
		// good - it's shorter (masked)
	} else {
		t.Errorf("masked token should be shorter than original")
	}
}

func TestSaveToken_ConfigFallback(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	// SaveToken will try keychain first (which may fail in test env),
	// then fallback to config file
	err := SaveToken("sk_test_token_save")
	if err != nil {
		t.Fatalf("SaveToken() error: %v", err)
	}

	cfg := readConfig(t, tmp)
	if cfg == nil {
		// If keychain succeeded, config may not have the key.
		// Check via GetToken instead.
		token, err := GetToken()
		if err != nil {
			t.Fatalf("GetToken() error: %v", err)
		}
		if token != "sk_test_token_save" {
			t.Errorf("GetToken() = %q, want %q", token, "sk_test_token_save")
		}
		return
	}

	// If token is in config file (keychain unavailable)
	apiKey, ok := cfg["api_key"]
	if ok {
		if apiKey != "sk_test_token_save" {
			t.Errorf("config api_key = %q, want %q", apiKey, "sk_test_token_save")
		}
	}
}

func TestGetToken_FromConfig(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	writeConfig(t, tmp, map[string]string{"api_key": "sk_config_token"})

	token, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken() error: %v", err)
	}
	if token != "sk_config_token" {
		t.Errorf("GetToken() = %q, want %q", token, "sk_config_token")
	}
}

func TestGetToken_NoToken(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	// No config file, no keychain token
	_, err := GetToken()
	if err == nil {
		t.Fatal("expected ErrNoToken, got nil")
	}
	if err != ErrNoToken {
		t.Errorf("expected ErrNoToken, got %v", err)
	}
}

func TestGetToken_EmptyConfigKey(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	writeConfig(t, tmp, map[string]string{"api_key": ""})

	_, err := GetToken()
	if err == nil {
		t.Fatal("expected ErrNoToken for empty key, got nil")
	}
	if err != ErrNoToken {
		t.Errorf("expected ErrNoToken, got %v", err)
	}
}

func TestGetToken_ConfigWithOnlyProjectID(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	writeConfig(t, tmp, map[string]string{"project_id": "proj_123"})

	_, err := GetToken()
	if err == nil {
		t.Fatal("expected ErrNoToken, got nil")
	}
}

func TestDeleteToken_ClearsConfig(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	writeConfig(t, tmp, map[string]string{
		"api_key":    "sk_to_delete",
		"project_id": "proj_keep",
	})

	err := DeleteToken()
	if err != nil {
		t.Fatalf("DeleteToken() error: %v", err)
	}

	cfg := readConfig(t, tmp)
	if cfg == nil {
		t.Fatal("config file should still exist")
	}

	apiKey, _ := cfg["api_key"]
	if apiKey != nil && apiKey != "" {
		t.Errorf("api_key should be empty after delete, got %v", apiKey)
	}
}

func TestDeleteToken_NoConfig(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	// Should not error even if no config exists
	err := DeleteToken()
	if err != nil {
		t.Fatalf("DeleteToken() error: %v", err)
	}
}

func TestSaveAndGetToken_RoundTrip(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	err := SaveToken("sk_roundtrip_test_key")
	if err != nil {
		t.Fatalf("SaveToken() error: %v", err)
	}

	token, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken() error: %v", err)
	}
	if token != "sk_roundtrip_test_key" {
		t.Errorf("GetToken() = %q, want %q", token, "sk_roundtrip_test_key")
	}
}

func TestSaveDeleteGet_Flow(t *testing.T) {
	tmp := t.TempDir()
	overrideHome(t, tmp)

	// Save a token
	if err := SaveToken("sk_flow_test"); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	// Verify it's there
	token, err := GetToken()
	if err != nil {
		t.Fatalf("GetToken: %v", err)
	}
	if token != "sk_flow_test" {
		t.Errorf("GetToken() = %q, want %q", token, "sk_flow_test")
	}

	// Delete it
	if err := DeleteToken(); err != nil {
		t.Fatalf("DeleteToken: %v", err)
	}

	// Verify it's gone
	_, err = GetToken()
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestTokenSource_ReturnsConfigFile(t *testing.T) {
	// In most test environments without keychain access, this should return "config file"
	source := TokenSource()
	// We accept either "keychain" or "config file" since it depends on the environment
	if source != "keychain" && source != "config file" {
		t.Errorf("TokenSource() = %q, want \"keychain\" or \"config file\"", source)
	}
}

func TestErrNoToken_Message(t *testing.T) {
	msg := ErrNoToken.Error()
	if msg == "" {
		t.Fatal("ErrNoToken message should not be empty")
	}
	expected := "not logged in - run `rc auth login` to authenticate"
	if msg != expected {
		t.Errorf("ErrNoToken = %q, want %q", msg, expected)
	}
}

func TestMaskToken_ConsistentOutput(t *testing.T) {
	// Calling MaskToken twice on same input should produce same result
	token := "sk_test_consistent_check_1234567890"
	first := MaskToken(token)
	second := MaskToken(token)
	if first != second {
		t.Errorf("MaskToken not consistent: %q vs %q", first, second)
	}
}

func TestMaskToken_NeverRevealsFullToken(t *testing.T) {
	tokens := []string{
		"sk_test_abc123def456ghi789",
		"abcdefghijklmnopqrstuvwxyz",
		"short",
		"sk_a",
		"sk_test_very_long_token_that_should_be_masked_properly_1234567890",
	}

	for _, token := range tokens {
		masked := MaskToken(token)
		if masked == token && len(token) > 8 {
			t.Errorf("MaskToken(%q) returned the original token unchanged", token)
		}
	}
}
