package auth_test

import (
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
