package charts_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestChartsOverviewTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "overview", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "revenue")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/metrics/overview")
}

func TestChartsOverviewJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "overview", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"overview_metrics\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/metrics/overview")
}

func TestChartsShowTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "show", "revenue", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Revenue")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/charts/revenue")
}

func TestChartsShowJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "show", "revenue", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"name\": \"revenue\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/charts/revenue")
}

func TestChartsShowMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "show"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestChartsOptionsTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "options", "revenue", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "country")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/charts/revenue/options")
}

func TestChartsOptionsJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "options", "revenue", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "country")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/charts/revenue/options")
}

func TestChartsOptionsMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "options"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestChartsOverviewWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "charts", "overview", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "revenue")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/metrics/overview")
}

func TestChartsOverviewProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "overview", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "revenue")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/metrics/overview")
}

func TestChartsOverviewNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "overview"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestChartsOverviewAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "overview"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestChartsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "overview", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestChartsOverviewHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "overview", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestChartsShowHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "show", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestChartsOptionsHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "options", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestChartsRootHelp(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "charts")
}

func TestChartsUnknownSubcommand(t *testing.T) {
	result := cmdtest.Run(t, []string{"charts", "nope"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "charts")
}
