package cmd_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestDoctorSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"doctor"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "API access:  OK")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestWhoamiJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"whoami", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"profile": "cmdtest"`)
	cmdtest.AssertOutputContains(t, result, `"default_project": "proj_cmdtest"`)
}

func TestInitSetsProfileProject(t *testing.T) {
	result := cmdtest.Run(t, []string{"init", "--profile-name", "staging", "--project", "proj_staging"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Initialized profile staging")
}

func TestConfigProfilesJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"config", "profiles", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"name": "cmdtest"`)
	cmdtest.AssertOutputContains(t, result, `"project_id": "proj_cmdtest"`)
}
