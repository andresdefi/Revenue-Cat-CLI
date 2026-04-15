package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewRootCmd_NotNil(t *testing.T) {
	root := NewRootCmd()
	if root == nil {
		t.Fatal("NewRootCmd() returned nil")
	}
}

func TestNewRootCmd_Use(t *testing.T) {
	root := NewRootCmd()
	if root.Use != "rc" {
		t.Errorf("Use = %q, want %q", root.Use, "rc")
	}
}

func TestNewRootCmd_Short(t *testing.T) {
	root := NewRootCmd()
	if root.Short == "" {
		t.Error("Short description should not be empty")
	}
	if !strings.Contains(root.Short, "RevenueCat") {
		t.Errorf("Short = %q, should contain 'RevenueCat'", root.Short)
	}
}

func TestNewRootCmd_Long(t *testing.T) {
	root := NewRootCmd()
	if root.Long == "" {
		t.Error("Long description should not be empty")
	}
	if !strings.Contains(root.Long, "RevenueCat") {
		t.Errorf("Long description should contain 'RevenueCat'")
	}
}

func TestNewRootCmd_HasExpectedSubcommands(t *testing.T) {
	root := NewRootCmd()

	expected := []string{
		"auth", "projects", "apps", "products", "entitlements",
		"offerings", "packages", "customers", "subscriptions",
		"purchases", "webhooks", "charts", "paywalls",
		"audit-logs", "collaborators", "currencies", "version",
		"completion", "init", "doctor", "whoami", "config",
		"mcp", "export", "import",
	}

	commands := make(map[string]bool)
	for _, cmd := range root.Commands() {
		commands[cmd.Name()] = true
	}

	for _, name := range expected {
		if !commands[name] {
			t.Errorf("root command missing subcommand: %s", name)
		}
	}
}

func TestNewRootCmd_SubcommandCount(t *testing.T) {
	root := NewRootCmd()
	commands := root.Commands()

	// meta/foundation + auth + project/product/customer/integration groups + mcp + transfer
	expectedMin := 25
	if len(commands) < expectedMin {
		t.Errorf("command count = %d, want >= %d", len(commands), expectedMin)
	}
}

func TestNewRootCmd_HasPersistentFlags(t *testing.T) {
	root := NewRootCmd()

	flags := []string{"project", "output"}
	for _, name := range flags {
		f := root.PersistentFlags().Lookup(name)
		if f == nil {
			t.Errorf("root command missing persistent flag: --%s", name)
		}
	}
}

func TestNewRootCmd_ShortFlags(t *testing.T) {
	root := NewRootCmd()

	tests := []struct {
		short string
		long  string
	}{
		{"p", "project"},
		{"o", "output"},
	}

	for _, tt := range tests {
		f := root.PersistentFlags().ShorthandLookup(tt.short)
		if f == nil {
			t.Errorf("root command missing short flag: -%s", tt.short)
		} else if f.Name != tt.long {
			t.Errorf("short flag -%s maps to %s, want %s", tt.short, f.Name, tt.long)
		}
	}
}

func TestNewRootCmd_SilencesUsage(t *testing.T) {
	root := NewRootCmd()
	if !root.SilenceUsage {
		t.Error("SilenceUsage should be true")
	}
}

func TestNewRootCmd_SilencesErrors(t *testing.T) {
	root := NewRootCmd()
	if !root.SilenceErrors {
		t.Error("SilenceErrors should be true")
	}
}

func TestNewRootCmd_SuggestionsEnabled(t *testing.T) {
	root := NewRootCmd()
	if root.SuggestionsMinimumDistance != 2 {
		t.Errorf("SuggestionsMinimumDistance = %d, want 2", root.SuggestionsMinimumDistance)
	}
}

func TestNewRootCmd_HelpContainsKeyText(t *testing.T) {
	root := NewRootCmd()
	help := root.Long

	keywords := []string{
		"RevenueCat",
		"API v2",
		"rc auth login",
		"rc projects list",
		"rc products list",
		"rc customers lookup",
		"rc charts overview",
	}

	for _, kw := range keywords {
		if !strings.Contains(help, kw) {
			t.Errorf("root help text missing keyword: %q", kw)
		}
	}
}

func TestNewRootCmd_HelpOutput(t *testing.T) {
	root := NewRootCmd()

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"--help"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("Execute --help error: %v", err)
	}

	helpOutput := buf.String()

	expectedTexts := []string{
		"rc",
		"auth login",
		"projects list",
		"products list",
	}

	for _, text := range expectedTexts {
		if !strings.Contains(helpOutput, text) {
			t.Errorf("help output should contain %q", text)
		}
	}
}

func TestNewRootCmd_HelpContainsFlags(t *testing.T) {
	root := NewRootCmd()

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"--help"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("Execute --help error: %v", err)
	}

	helpOutput := buf.String()

	if !strings.Contains(helpOutput, "--project") {
		t.Error("help should mention --project flag")
	}
	if !strings.Contains(helpOutput, "--output") {
		t.Error("help should mention --output flag")
	}
}

func TestNewRootCmd_HelpContainsSubcommands(t *testing.T) {
	root := NewRootCmd()

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"--help"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("Execute --help error: %v", err)
	}

	helpOutput := buf.String()

	subcommands := []string{
		"version",
		"auth",
		"projects",
		"products",
		"customers",
		"subscriptions",
		"charts",
		"webhooks",
		"offerings",
	}

	for _, sub := range subcommands {
		if !strings.Contains(helpOutput, sub) {
			t.Errorf("help output should list %q subcommand", sub)
		}
	}
}

func TestNewRootCmd_Aliases(t *testing.T) {
	root := NewRootCmd()

	aliasTests := map[string][]string{
		"products":      {"product", "prod"},
		"entitlements":  {"entitlement", "ent"},
		"offerings":     {"offering", "off"},
		"packages":      {"package", "pkg"},
		"customers":     {"customer", "cust"},
		"subscriptions": {"subscription", "sub"},
		"currencies":    {"currency", "vc"},
		"webhooks":      {"webhook", "wh"},
		"charts":        {"chart", "metrics"},
		"purchases":     {"purchase"},
		"paywalls":      {"paywall"},
		"audit-logs":    {"audit", "logs"},
		"collaborators": {"collaborator", "collab"},
		"apps":          {"app"},
		"projects":      {"project", "proj"},
	}

	for _, cmd := range root.Commands() {
		if expected, ok := aliasTests[cmd.Name()]; ok {
			for _, alias := range expected {
				found := false
				for _, a := range cmd.Aliases {
					if a == alias {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("command %s missing alias: %s", cmd.Name(), alias)
				}
			}
		}
	}
}

func TestNewRootCmd_AuthSubcommands(t *testing.T) {
	root := NewRootCmd()
	authCmd, _, err := root.Find([]string{"auth"})
	if err != nil {
		t.Fatalf("Find auth: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range authCmd.Commands() {
		subNames[c.Name()] = true
	}

	for _, name := range []string{"login", "status", "logout", "doctor", "validate"} {
		if !subNames[name] {
			t.Errorf("auth should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_ProjectsSubcommands(t *testing.T) {
	root := NewRootCmd()
	projCmd, _, err := root.Find([]string{"projects"})
	if err != nil {
		t.Fatalf("Find projects: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range projCmd.Commands() {
		subNames[c.Name()] = true
	}

	for _, name := range []string{"list", "create", "doctor", "set-default"} {
		if !subNames[name] {
			t.Errorf("projects should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_ProductsSubcommands(t *testing.T) {
	root := NewRootCmd()
	prodCmd, _, err := root.Find([]string{"products"})
	if err != nil {
		t.Fatalf("Find products: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range prodCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list", "get", "create", "update", "delete", "archive", "unarchive", "push-to-store"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("products should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_CustomersSubcommands(t *testing.T) {
	root := NewRootCmd()
	custCmd, _, err := root.Find([]string{"customers"})
	if err != nil {
		t.Fatalf("Find customers: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range custCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list", "lookup", "create", "delete", "entitlements", "subscriptions", "purchases", "aliases", "attributes", "set-attributes", "grant", "revoke", "assign-offering", "transfer", "restore-purchase", "invoices", "invoice-file"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("customers should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_SubscriptionsSubcommands(t *testing.T) {
	root := NewRootCmd()
	subCmd, _, err := root.Find([]string{"subscriptions"})
	if err != nil {
		t.Fatalf("Find subscriptions: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range subCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list", "get", "transactions", "entitlements", "cancel", "refund", "refund-transaction", "management-url"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("subscriptions should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_EntitlementsSubcommands(t *testing.T) {
	root := NewRootCmd()
	entCmd, _, err := root.Find([]string{"entitlements"})
	if err != nil {
		t.Fatalf("Find entitlements: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range entCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list", "get", "create", "update", "delete", "archive", "unarchive", "products", "attach", "detach"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("entitlements should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_OfferingsSubcommands(t *testing.T) {
	root := NewRootCmd()
	offCmd, _, err := root.Find([]string{"offerings"})
	if err != nil {
		t.Fatalf("Find offerings: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range offCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list", "get", "create", "update", "delete", "archive", "unarchive"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("offerings should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_PackagesSubcommands(t *testing.T) {
	root := NewRootCmd()
	pkgCmd, _, err := root.Find([]string{"packages"})
	if err != nil {
		t.Fatalf("Find packages: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range pkgCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list", "get", "create", "update", "delete", "products", "attach", "detach"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("packages should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_AppsSubcommands(t *testing.T) {
	root := NewRootCmd()
	appsCmd, _, err := root.Find([]string{"apps"})
	if err != nil {
		t.Fatalf("Find apps: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range appsCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list", "get", "create", "update", "delete", "public-keys", "storekit-config"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("apps should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_WebhooksSubcommands(t *testing.T) {
	root := NewRootCmd()
	whCmd, _, err := root.Find([]string{"webhooks"})
	if err != nil {
		t.Fatalf("Find webhooks: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range whCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list", "get", "create", "update", "delete"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("webhooks should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_ChartsSubcommands(t *testing.T) {
	root := NewRootCmd()
	chartsCmd, _, err := root.Find([]string{"charts"})
	if err != nil {
		t.Fatalf("Find charts: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range chartsCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"overview", "show", "options"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("charts should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_PurchasesSubcommands(t *testing.T) {
	root := NewRootCmd()
	purchCmd, _, err := root.Find([]string{"purchases"})
	if err != nil {
		t.Fatalf("Find purchases: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range purchCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list", "get", "entitlements", "refund"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("purchases should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_PaywallsSubcommands(t *testing.T) {
	root := NewRootCmd()
	pwCmd, _, err := root.Find([]string{"paywalls"})
	if err != nil {
		t.Fatalf("Find paywalls: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range pwCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list", "get", "create", "delete"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("paywalls should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_AuditLogsSubcommands(t *testing.T) {
	root := NewRootCmd()
	auditCmd, _, err := root.Find([]string{"audit-logs"})
	if err != nil {
		t.Fatalf("Find audit-logs: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range auditCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("audit-logs should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_CollaboratorsSubcommands(t *testing.T) {
	root := NewRootCmd()
	collabCmd, _, err := root.Find([]string{"collaborators"})
	if err != nil {
		t.Fatalf("Find collaborators: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range collabCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("collaborators should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_CurrenciesSubcommands(t *testing.T) {
	root := NewRootCmd()
	currCmd, _, err := root.Find([]string{"currencies"})
	if err != nil {
		t.Fatalf("Find currencies: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range currCmd.Commands() {
		subNames[c.Name()] = true
	}

	expected := []string{"list", "get", "create", "update", "delete", "archive", "unarchive", "balance", "credit", "set-balance"}
	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("currencies should have subcommand %q", name)
		}
	}
}

func TestNewRootCmd_UnknownCommand(t *testing.T) {
	root := NewRootCmd()

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"nonexistent"})

	err := root.Execute()
	if err == nil {
		t.Error("expected error for unknown command")
	}
}

func TestNewRootCmd_NoArgsShowsHelp(t *testing.T) {
	root := NewRootCmd()

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{})

	err := root.Execute()
	if err != nil {
		t.Fatalf("Execute with no args error: %v", err)
	}
}

// --- v0.2.0 tests ---

func TestNewRootCmd_MCPSubcommand(t *testing.T) {
	root := NewRootCmd()

	commands := make(map[string]bool)
	for _, cmd := range root.Commands() {
		commands[cmd.Name()] = true
	}

	if !commands["mcp"] {
		t.Error("root command missing 'mcp' subcommand")
	}
}

func TestNewRootCmd_MCPHasServe(t *testing.T) {
	root := NewRootCmd()
	mcpCmd, _, err := root.Find([]string{"mcp"})
	if err != nil {
		t.Fatalf("Find mcp: %v", err)
	}

	subNames := make(map[string]bool)
	for _, c := range mcpCmd.Commands() {
		subNames[c.Name()] = true
	}

	if !subNames["serve"] {
		t.Error("mcp should have 'serve' subcommand")
	}
}

func TestNewRootCmd_ExportSubcommand(t *testing.T) {
	root := NewRootCmd()

	commands := make(map[string]bool)
	for _, cmd := range root.Commands() {
		commands[cmd.Name()] = true
	}

	if !commands["export"] {
		t.Error("root command missing 'export' subcommand")
	}
}

func TestNewRootCmd_ImportSubcommand(t *testing.T) {
	root := NewRootCmd()

	commands := make(map[string]bool)
	for _, cmd := range root.Commands() {
		commands[cmd.Name()] = true
	}

	if !commands["import"] {
		t.Error("root command missing 'import' subcommand")
	}
}

func TestNewRootCmd_ProfileFlag(t *testing.T) {
	root := NewRootCmd()

	f := root.PersistentFlags().Lookup("profile")
	if f == nil {
		t.Fatal("root command missing --profile persistent flag")
	}
	if f.DefValue != "" {
		t.Errorf("--profile default = %q, want empty string", f.DefValue)
	}
}

func TestNewRootCmd_ExportHasFileFlag(t *testing.T) {
	root := NewRootCmd()

	var exportCmd *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Name() == "export" {
			exportCmd = cmd
			break
		}
	}
	if exportCmd == nil {
		t.Fatal("export subcommand not found")
	}

	f := exportCmd.Flags().Lookup("file")
	if f == nil {
		t.Error("export command missing --file flag")
	}
}

func TestNewRootCmd_ImportHasFileFlag(t *testing.T) {
	root := NewRootCmd()

	var importCmd *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Name() == "import" {
			importCmd = cmd
			break
		}
	}
	if importCmd == nil {
		t.Fatal("import subcommand not found")
	}

	f := importCmd.Flags().Lookup("file")
	if f == nil {
		t.Error("import command missing --file flag")
	}
}

func TestNewRootCmd_HelpContainsNewSubcommands(t *testing.T) {
	root := NewRootCmd()

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"--help"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("Execute --help error: %v", err)
	}

	helpOutput := buf.String()

	newSubs := []string{"mcp", "export", "import"}
	for _, sub := range newSubs {
		if !strings.Contains(helpOutput, sub) {
			t.Errorf("help output should list %q subcommand", sub)
		}
	}
}

func TestNewRootCmd_HelpContainsProfileFlag(t *testing.T) {
	root := NewRootCmd()

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"--help"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("Execute --help error: %v", err)
	}

	if !strings.Contains(buf.String(), "--profile") {
		t.Error("help should mention --profile flag")
	}
}
