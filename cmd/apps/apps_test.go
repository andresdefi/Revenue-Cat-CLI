package apps_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestAppsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "app_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps")
}

func TestAppsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps")
}

func TestAppsListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "apps", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "app_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps")
}

func TestAppsListProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "app_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/apps")
}

func TestAppsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestAppsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestAppsGetTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "get", "app_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "app_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps/app_cmdtest")
}

func TestAppsGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "get", "app_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"id\": \"app_cmdtest\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps/app_cmdtest")
}

func TestAppsGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestAppsCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "create", "--name", "iOS App", "--type", "app_store", "--bundle-id", "com.example.app", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "app_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/apps")
}

func TestAppsCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "create"})
	cmdtest.AssertErrorContains(t, result, "missing required value: App name")
}

func TestAppsCreateMissingTypeFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "create", "--name", "iOS App"})
	cmdtest.AssertErrorContains(t, result, "missing required value: Platform type")
}

func TestAppsDeleteSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "delete", "app_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "deleted")
	cmdtest.AssertRequested(t, result, "DELETE", "/projects/proj_cmdtest/apps/app_cmdtest")
}

func TestAppsDeleteMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "delete"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestAppsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestAppsHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestAppsUpdateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "update", "app_cmdtest", "--name", "Renamed App", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "app_cmdtest")
	cmdtest.AssertRequestJSON(t, result, "POST", "/projects/proj_cmdtest/apps/app_cmdtest", map[string]any{
		"name": "Renamed App",
	})
}

func TestAppsUpdateAppStoreCredentials(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "SubscriptionKey_ABC123.p8")
	keyContents := "-----BEGIN PRIVATE KEY-----\nfixture-key\n-----END PRIVATE KEY-----\n"
	if err := os.WriteFile(keyPath, []byte(keyContents), 0o600); err != nil {
		t.Fatalf("write subscription key fixture: %v", err)
	}

	result := cmdtest.Run(t, []string{
		"apps", "update", "app_cmdtest",
		"--shared-secret", "1234567890abcdef1234567890abcdef",
		"--subscription-key-file", keyPath,
		"--subscription-key-id", "ABC123",
		"--subscription-key-issuer", "5a049d62-1b9b-453c-b605-1988189d8129",
		"--output", "json",
	})

	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequestJSON(t, result, "POST", "/projects/proj_cmdtest/apps/app_cmdtest", map[string]any{
		"app_store": map[string]any{
			"shared_secret":            "1234567890abcdef1234567890abcdef",
			"subscription_private_key": keyContents,
			"subscription_key_id":      "ABC123",
			"subscription_key_issuer":  "5a049d62-1b9b-453c-b605-1988189d8129",
		},
	})
}

func TestAppsUpdateAppStoreConnectCredentials(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "AuthKey_ABC123.p8")
	keyContents := "-----BEGIN PRIVATE KEY-----\nconnect-key\n-----END PRIVATE KEY-----\n"
	if err := os.WriteFile(keyPath, []byte(keyContents), 0o600); err != nil {
		t.Fatalf("write App Store Connect key fixture: %v", err)
	}

	result := cmdtest.Run(t, []string{
		"apps", "update", "app_cmdtest",
		"--app-store-connect-api-key-file", keyPath,
		"--app-store-connect-api-key-id", "ABC123",
		"--app-store-connect-api-key-issuer", "5a049d62-1b9b-453c-b605-1988189d8129",
		"--app-store-connect-vendor-number", "12345678",
		"--output", "json",
	})

	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequestJSON(t, result, "POST", "/projects/proj_cmdtest/apps/app_cmdtest", map[string]any{
		"app_store": map[string]any{
			"app_store_connect_api_key":        keyContents,
			"app_store_connect_api_key_id":     "ABC123",
			"app_store_connect_api_key_issuer": "5a049d62-1b9b-453c-b605-1988189d8129",
			"app_store_connect_vendor_number":  "12345678",
		},
	})
}

func TestAppsUpdateSubscriptionKeyFileError(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.p8")
	result := cmdtest.Run(t, []string{"apps", "update", "app_cmdtest", "--subscription-key-file", missingPath})
	cmdtest.AssertErrorContains(t, result, "read subscription key file")
}

func TestAppsUpdateAppStoreConnectKeyFileError(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.p8")
	result := cmdtest.Run(t, []string{"apps", "update", "app_cmdtest", "--app-store-connect-api-key-file", missingPath})
	cmdtest.AssertErrorContains(t, result, "read App Store Connect API key file")
}

func TestAppsUpdatePlayStoreServiceAccountFileUnsupported(t *testing.T) {
	credentialsPath := filepath.Join(t.TempDir(), "service-account.json")
	if err := os.WriteFile(credentialsPath, []byte(`{"client_email":"fixture@example.com"}`), 0o600); err != nil {
		t.Fatalf("write service account fixture: %v", err)
	}

	result := cmdtest.Run(t, []string{"apps", "update", "app_cmdtest", "--service-account-file", credentialsPath})
	cmdtest.AssertErrorContains(t, result, "does not document a Play Store service-account credential field")
}

func TestAppsUpdateMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "update"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestAppsUpdateAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"apps", "update", "app_cmdtest", "--name", "Renamed App", "--output", "json"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}
