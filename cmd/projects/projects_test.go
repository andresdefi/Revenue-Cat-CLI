package projects_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestProjectsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "proj_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestProjectsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestProjectsListAll(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "--all", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "proj_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestProjectsListLimit(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "--limit", "1", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "proj_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestProjectsListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "projects", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "proj_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestProjectsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestProjectsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestProjectsCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "create", "--name", "Command Test Project", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "proj_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects")
}

func TestProjectsCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "create"})
	cmdtest.AssertErrorContains(t, result, "required flag")
}

func TestProjectsSetDefaultSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "set-default", "proj_new_default"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Default project set")
}

func TestProjectsSetDefaultMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "set-default"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestProjectsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestProjectsListHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestProjectsCreateHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "create", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestProjectsSetDefaultHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "set-default", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestProjectsRootHelp(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "projects")
}

func TestProjectsUnknownSubcommand(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "nope"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "projects")
}

func TestProjectsListShortOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"projects", "list", "-o", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "proj_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}
