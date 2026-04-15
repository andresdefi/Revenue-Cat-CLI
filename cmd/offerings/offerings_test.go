package offerings_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestOfferingsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "ofrnge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings")
}

func TestOfferingsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings")
}

func TestOfferingsListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "offerings", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "ofrnge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings")
}

func TestOfferingsListProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "ofrnge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/offerings")
}

func TestOfferingsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestOfferingsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestOfferingsGetTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "get", "ofrnge_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "ofrnge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest")
}

func TestOfferingsGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "get", "ofrnge_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"id\": \"ofrnge_cmdtest\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest")
}

func TestOfferingsGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestOfferingsCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "create", "--lookup-key", "default", "--display-name", "Default", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "ofrnge_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/offerings")
}

func TestOfferingsCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "create"})
	cmdtest.AssertErrorContains(t, result, "required")
}

func TestOfferingsDeleteSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "delete", "ofrnge_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "deleted")
	cmdtest.AssertRequested(t, result, "DELETE", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest")
}

func TestOfferingsDeleteMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "delete"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestOfferingsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestOfferingsHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestOfferingsUpdateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "update", "ofrnge_cmdtest", "--display-name", "Default Plus", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "ofrnge_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest")
}

func TestOfferingsUpdateMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "update"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestOfferingsUpdateAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "update", "ofrnge_cmdtest", "--display-name", "Default Plus", "--output", "json"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}
