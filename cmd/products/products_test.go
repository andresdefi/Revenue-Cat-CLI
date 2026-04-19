package products_test

import (
	"net/http"
	"strings"
	"testing"

	productcmd "github.com/andresdefi/rc/cmd/products"
	"github.com/andresdefi/rc/internal/cmdtest"
	"github.com/andresdefi/rc/internal/cmdutil"
)

func TestProductsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "prod_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/products")
}

func TestProductsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"object": "list"`)
}

func TestProductsAgentModeUsesCompactJSONDefaultFields(t *testing.T) {
	result := cmdtest.Run(t, []string{"--agent", "products", "list"})
	cmdtest.AssertSuccess(t, result)
	if !strings.Contains(result.Stdout, `"items":[`) {
		t.Fatalf("agent output should be compact JSON, got stdout:\n%s", result.Stdout)
	}
	cmdtest.AssertOutputContains(t, result, `"store_identifier":"com.example.premium.monthly"`)
	if strings.Contains(result.Stdout, "created_at") {
		t.Fatalf("agent output should use the default product preset, got stdout:\n%s", result.Stdout)
	}
	if strings.Contains(result.Stderr, "next:") {
		t.Fatalf("agent output should suppress hints, got stderr:\n%s", result.Stderr)
	}
}

func TestProductsAgentEnvUsesCompactJSONDefaultFields(t *testing.T) {
	t.Setenv("RC_AGENT", "1")
	result := cmdtest.Run(t, []string{"products", "list"})
	cmdtest.AssertSuccess(t, result)
	if !strings.Contains(result.Stdout, `"items":[`) {
		t.Fatalf("RC_AGENT output should be compact JSON, got stdout:\n%s", result.Stdout)
	}
	if strings.Contains(result.Stdout, "created_at") {
		t.Fatalf("RC_AGENT output should use the default product preset, got stdout:\n%s", result.Stdout)
	}
}

func TestProductsAgentModeExplicitOutputWins(t *testing.T) {
	result := cmdtest.Run(t, []string{"--agent", "--output", "table", "products", "list"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "STORE ID")
	if strings.Contains(result.Stdout, `"items"`) {
		t.Fatalf("explicit table output should beat --agent, got stdout:\n%s", result.Stdout)
	}
}

func TestProductsAgentModeExplicitFieldsWin(t *testing.T) {
	result := cmdtest.Run(t, []string{"--agent", "--fields", "id", "products", "list"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"id":"prod_cmdtest"`)
	if strings.Contains(result.Stdout, "store_identifier") {
		t.Fatalf("explicit --fields should beat --agent default preset, got stdout:\n%s", result.Stdout)
	}
}

func TestProductsAgentModeDoesNotWarnWhenMutationHasNoPreset(t *testing.T) {
	result := cmdtest.Run(t, []string{"--agent", "products", "create", "--store-id", "com.example.premium.monthly", "--app-id", "app_cmdtest", "--type", "subscription"})
	cmdtest.AssertSuccess(t, result)
	if strings.Contains(result.Stderr, "no default preset") {
		t.Fatalf("agent mutation should not warn about missing presets, got stderr:\n%s", result.Stderr)
	}
	if strings.Contains(result.Stderr, "next:") {
		t.Fatalf("agent mutation should suppress hints, got stderr:\n%s", result.Stderr)
	}
}

func TestProductsListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", cmdtest.TestProfile, "products", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "prod_cmdtest")
}

func TestProductsListProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/products")
}

func TestProductsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestProductsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list"}, cmdtest.WithAPIError(400, "parameter_error", "bad products request"))
	cmdtest.AssertErrorContains(t, result, "bad products request")
}

func TestProductsGetTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "get", "prod_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Premium Monthly")
}

func TestProductsGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "get", "prod_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"id": "prod_cmdtest"`)
}

func TestProductsGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestProductsCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "create", "--store-id", "com.example.premium.monthly", "--app-id", "app_cmdtest", "--type", "subscription", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Product created successfully")
	cmdtest.AssertOutputContains(t, result, "next: rc products push-to-store prod_cmdtest")
}

func TestProductsFieldsPresetRegistration(t *testing.T) {
	projectID := "proj_cmdtest"
	outputFormat := ""
	root := productcmd.NewProductsCmd(&projectID, &outputFormat)
	want := "id,store_identifier,type,state,display_name,app_id"

	for _, name := range []string{"list", "get"} {
		t.Run(name, func(t *testing.T) {
			cmd, _, err := root.Find([]string{name})
			if err != nil {
				t.Fatalf("find %s: %v", name, err)
			}
			if got := cmdutil.FieldsPreset(cmd); got != want {
				t.Fatalf("preset for %s = %q, want %q", name, got, want)
			}
		})
	}
}

func TestProductsCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "create", "--store-id", "com.example.premium.monthly"})
	cmdtest.AssertErrorContains(t, result, "missing required value")
}

func TestProductsUpdateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "update", "prod_cmdtest", "--display-name", "Premium Annual", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Product updated")
}

func TestProductsUpdateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "update", "prod_cmdtest"})
	cmdtest.AssertErrorContains(t, result, "required flag")
}

func TestProductsDeleteSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "delete", "prod_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "deleted")
}

func TestProductsDeleteMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "delete"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestProductsArchiveSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "archive", "prod_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "archived")
}

func TestProductsUnarchiveSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "unarchive", "prod_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "unarchived")
}

func TestProductsPushToStoreIAPSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "push-to-store", "prod_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "pushed to store")
	cmdtest.AssertRequestBody(t, result, http.MethodPost, "/projects/proj_cmdtest/products/prod_cmdtest/create_in_store", "")
}

func TestProductsPushToStoreSubscriptionBody(t *testing.T) {
	result := cmdtest.Run(t, []string{
		"products", "push-to-store", "prod_cmdtest",
		"--subscription-duration", "ONE_MONTH",
		"--subscription-group-name", "Premium Subscriptions",
		"--subscription-group-id", "sub_group_123",
	})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequestJSON(t, result, http.MethodPost, "/projects/proj_cmdtest/products/prod_cmdtest/create_in_store", map[string]any{
		"store_information": map[string]any{
			"duration":                "ONE_MONTH",
			"subscription_group_name": "Premium Subscriptions",
			"subscription_group_id":   "sub_group_123",
		},
	})
}

func TestProductsPushToStoreSubscriptionMissingDuration(t *testing.T) {
	result := cmdtest.Run(t, []string{
		"products", "push-to-store", "prod_cmdtest",
		"--subscription-group-name", "Premium Subscriptions",
	})
	cmdtest.AssertErrorContains(t, result, "--subscription-duration is required")
	cmdtest.AssertNotRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/products/prod_cmdtest/create_in_store")
}

func TestProductsPushToStoreSubscriptionInvalidDuration(t *testing.T) {
	result := cmdtest.Run(t, []string{
		"products", "push-to-store", "prod_cmdtest",
		"--subscription-duration", "P1M",
		"--subscription-group-name", "Premium Subscriptions",
	})
	cmdtest.AssertErrorContains(t, result, "invalid --subscription-duration")
	cmdtest.AssertNotRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/products/prod_cmdtest/create_in_store")
}

func TestProductsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestProductsHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"products", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}
