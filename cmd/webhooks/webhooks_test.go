package webhooks_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestWebhooksListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "wh_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/integrations/webhooks")
}

func TestWebhooksListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/integrations/webhooks")
}

func TestWebhooksListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "webhooks", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "wh_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/integrations/webhooks")
}

func TestWebhooksListProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "list", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "wh_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/integrations/webhooks")
}

func TestWebhooksListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestWebhooksListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestWebhooksGetTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "get", "wh_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "wh_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/integrations/webhooks/wh_cmdtest")
}

func TestWebhooksGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "get", "wh_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"id\": \"wh_cmdtest\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/integrations/webhooks/wh_cmdtest")
}

func TestWebhooksGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestWebhooksCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "create", "--name", "Events", "--url", "https://example.com/revenuecat", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "wh_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/integrations/webhooks")
}

func TestWebhooksCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "create"})
	cmdtest.AssertErrorContains(t, result, "missing required value: Webhook name")
}

func TestWebhooksCreateMissingURLFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "create", "--name", "Events"})
	cmdtest.AssertErrorContains(t, result, "missing required value: Webhook URL")
}

func TestWebhooksDeleteSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "delete", "wh_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "deleted")
	cmdtest.AssertRequested(t, result, "DELETE", "/projects/proj_cmdtest/integrations/webhooks/wh_cmdtest")
}

func TestWebhooksDeleteMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "delete"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestWebhooksInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestWebhooksHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestWebhooksUpdateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "update", "wh_cmdtest", "--name", "Events v2", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "wh_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/integrations/webhooks/wh_cmdtest")
}

func TestWebhooksUpdateMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "update"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestWebhooksUpdateAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"webhooks", "update", "wh_cmdtest", "--name", "Events v2", "--output", "json"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}
