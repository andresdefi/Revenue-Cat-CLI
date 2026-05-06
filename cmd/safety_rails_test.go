package cmd_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestRequireProfileRejectsWrongProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--require-profile", "production", "projects", "list"})
	cmdtest.AssertErrorContains(t, result, `profile safety check failed`)
	cmdtest.AssertNotRequested(t, result, "GET", "/projects")
}

func TestFailIfProjectNameNotRejectsMismatch(t *testing.T) {
	result := cmdtest.Run(t, []string{"--fail-if-project-name-not", "Acme Production", "products", "list"})
	cmdtest.AssertErrorContains(t, result, `project safety check failed`)
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest")
	cmdtest.AssertNotRequested(t, result, "GET", "/projects/proj_cmdtest/products")
}

func TestFailIfProjectNameNotAllowsMatch(t *testing.T) {
	result := cmdtest.Run(t, []string{"--fail-if-project-name-not", "Command Test Project", "products", "list"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/products")
}

func TestRequireConfirmEnvOverridesYes(t *testing.T) {
	t.Setenv("RC_REQUIRE_CONFIRM", "1")

	result := cmdtest.Run(t, []string{"products", "delete", "prod_cmdtest", "--yes"})
	cmdtest.AssertErrorContains(t, result, `destructive operation requires typed confirmation`)
	cmdtest.AssertNotRequested(t, result, "DELETE", "/projects/proj_cmdtest/products/prod_cmdtest")
}
