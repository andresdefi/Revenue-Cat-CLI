package cmd

import (
	"strings"
	"testing"
)

func TestNewRootCmd_HasExpectedSubcommands(t *testing.T) {
	root := NewRootCmd()

	expected := []string{
		"auth", "projects", "apps", "products", "entitlements",
		"offerings", "packages", "customers", "subscriptions",
		"purchases", "webhooks", "charts", "paywalls",
		"audit-logs", "collaborators", "currencies", "version",
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

func TestNewRootCmd_HelpContainsKeyText(t *testing.T) {
	root := NewRootCmd()
	help := root.Long

	keywords := []string{
		"RevenueCat",
		"API v2",
		"rc auth login",
		"rc projects list",
	}

	for _, kw := range keywords {
		if !strings.Contains(help, kw) {
			t.Errorf("root help text missing keyword: %q", kw)
		}
	}
}

func TestNewRootCmd_SuggestionsEnabled(t *testing.T) {
	root := NewRootCmd()
	if root.SuggestionsMinimumDistance != 2 {
		t.Errorf("SuggestionsMinimumDistance = %d, want 2", root.SuggestionsMinimumDistance)
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
