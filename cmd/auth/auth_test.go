package auth_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestAuthStatusLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "status"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Logged in")
}

func TestAuthStatusProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "auth", "status"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "cmdtest")
}

func TestAuthStatusShowsProfileCount(t *testing.T) {
	result := cmdtest.Run(t,
		[]string{"auth", "status"},
		cmdtest.WithProfiles(map[string]cmdtest.ProfileConfig{
			"cmdtest": {ProjectID: cmdtest.TestProjectID, Token: cmdtest.TestToken},
			"staging": {ProjectID: "proj_staging", Token: "sk_staging_token"},
		}),
	)
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Profiles: 2 stored")
}

func TestAuthStatusNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "status"}, cmdtest.WithoutToken())
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Not logged in")
}

func TestAuthDoctorLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "doctor"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "API access:  OK")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestAuthValidateLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "validate"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "API access:  OK")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestAuthDoctorProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "auth", "doctor"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Profile:")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestAuthDoctorAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "doctor"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "API access:  failed")
}

func TestAuthDoctorNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "doctor"}, cmdtest.WithoutToken())
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Token:       not found")
}

func TestAuthLogoutSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "logout"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Logged out")
}

func TestAuthLogoutProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "auth", "logout"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "cmdtest")
}

func TestAuthLoginSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "login"}, cmdtest.WithStdin("sk_login_token\n"))
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Logged in successfully")
}

func TestAuthLoginEmptyToken(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "login"}, cmdtest.WithStdin("\n"))
	cmdtest.AssertErrorContains(t, result, "API key cannot be empty")
}

func TestAuthLoginRejectsInvalidPrefix(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "login"}, cmdtest.WithStdin("rk_test_token\n"))
	cmdtest.AssertErrorContains(t, result, "invalid API key prefix")
}

func TestAuthLoginProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "auth", "login"}, cmdtest.WithStdin("sk_login_token\n"))
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "cmdtest")
}

func TestAuthAddProjectInfersProfileName(t *testing.T) {
	result := cmdtest.Run(t,
		[]string{"auth", "add-project", "--key", "sk_impostor_token"},
		cmdtest.WithAcceptedTokens("sk_impostor_token"),
		cmdtest.WithHandler(func(w http.ResponseWriter, r *http.Request) {
			writeAuthTestJSON(w, http.StatusOK, map[string]any{
				"object":    "list",
				"items":     []any{map[string]any{"object": "project", "id": "projb26c2f72", "name": "Impostor", "created_at": 1776240000000}},
				"next_page": nil,
			})
		}),
	)
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Added project 'Impostor' (projb26c2f72) under profile 'impostor'")
	cmdtest.AssertRequested(t, result, "GET", "/projects")
}

func TestAuthAddProjectRequiresNameForMultipleProjects(t *testing.T) {
	result := cmdtest.Run(t,
		[]string{"auth", "add-project", "--key", "sk_multi_token"},
		cmdtest.WithAcceptedTokens("sk_multi_token"),
		cmdtest.WithHandler(func(w http.ResponseWriter, r *http.Request) {
			writeAuthTestJSON(w, http.StatusOK, map[string]any{
				"object": "list",
				"items": []any{
					map[string]any{"object": "project", "id": "proj_one", "name": "One", "created_at": 1776240000000},
					map[string]any{"object": "project", "id": "proj_two", "name": "Two", "created_at": 1776240000000},
				},
				"next_page": nil,
			})
		}),
	)
	cmdtest.AssertErrorContains(t, result, "pass --name")
}

func TestAuthStatusHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "status", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestAuthLoginHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "login", "--help"}, cmdtest.WithStdin("sk_login_token\n"))
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestAuthDoctorHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "doctor", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestAuthRootHelp(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Log in, check status")
}

func TestAuthUnknownSubcommand(t *testing.T) {
	result := cmdtest.Run(t, []string{"auth", "nope"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Log in, check status")
}

func writeAuthTestJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
