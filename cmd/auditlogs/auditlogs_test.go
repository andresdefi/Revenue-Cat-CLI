package auditlogs_test

import (
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestAuditLogsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "audit_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/audit_logs")
}

func TestAuditLogsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/audit_logs")
}

func TestAuditLogsListAll(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list", "--all", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "audit_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/audit_logs")
}

func TestAuditLogsListLimit(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list", "--limit", "1", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "audit_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/audit_logs")
}

func TestAuditLogsListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "audit-logs", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "audit_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/audit_logs")
}

func TestAuditLogsListProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "audit_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/audit_logs")
}

func TestAuditLogsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestAuditLogsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestAuditLogsListMissingProject(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list"}, cmdtest.WithoutProject())
	cmdtest.AssertErrorContains(t, result, "no project specified")
}

func TestAuditLogsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestAuditLogsHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestAuditLogsRootHelp(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "audit-logs")
}

func TestAuditLogsUnknownSubcommand(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "nope"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "audit-logs")
}

func TestAuditLogsListShortOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list", "-o", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "audit_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/audit_logs")
}

func TestAuditLogsListWithProjectShortFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list", "-p", "proj_override", "-o", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "audit_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/audit_logs")
}

func TestAuditLogsListLimitZero(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list", "--limit", "0", "-o", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "audit_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/audit_logs")
}

func TestAuditLogsListAllTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list", "--all", "-o", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "audit_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/audit_logs")
}

func TestAuditLogsListDateFlags(t *testing.T) {
	result := cmdtest.Run(t, []string{"audit-logs", "list", "-o", "json", "--start-date", "2026-04-01", "--end-date", "2026-04-14"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "audit_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/audit_logs")
}
