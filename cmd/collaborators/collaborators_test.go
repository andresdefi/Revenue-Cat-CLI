package collaborators_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestCollaboratorsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "collab_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/collaborators")
}

func TestCollaboratorsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/collaborators")
}

func TestCollaboratorsListAll(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list", "--all", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "collab_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/collaborators")
}

func TestCollaboratorsListLimit(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list", "--limit", "1", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "collab_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/collaborators")
}

func TestCollaboratorsListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "collaborators", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "collab_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/collaborators")
}

func TestCollaboratorsListProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "collab_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/collaborators")
}

func TestCollaboratorsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestCollaboratorsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestCollaboratorsListMissingProject(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list"}, cmdtest.WithoutProject())
	cmdtest.AssertErrorContains(t, result, "no project specified")
}

func TestCollaboratorsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestCollaboratorsHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestCollaboratorsRootHelp(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "collaborators")
}

func TestCollaboratorsUnknownSubcommand(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "nope"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "collaborators")
}

func TestCollaboratorsListShortOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list", "-o", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "collab_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/collaborators")
}

func TestCollaboratorsListWithProjectShortFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list", "-p", "proj_override", "-o", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "collab_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/collaborators")
}

func TestCollaboratorsListLimitZero(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list", "--limit", "0", "-o", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "collab_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/collaborators")
}

func TestCollaboratorsListAllTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list", "--all", "-o", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "collab_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/collaborators")
}

func TestCollaboratorsListDateFlags(t *testing.T) {
	result := cmdtest.Run(t, []string{"collaborators", "list", "-o", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "collab_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/collaborators")
}
