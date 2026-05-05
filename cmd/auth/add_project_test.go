package auth

import (
	"testing"

	"github.com/andresdefi/rc/internal/api"
)

func TestInferProfileName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "Impostor", want: "impostor"},
		{name: "My Great App!", want: "my-great-app"},
		{name: "RevenueCat: iOS + Android", want: "revenuecat-ios-android"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferProfileName(api.Project{ID: "proj_fallback", Name: tt.name})
			if got != tt.want {
				t.Fatalf("inferProfileName(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestInferProfileNameFallsBackToProjectID(t *testing.T) {
	got := inferProfileName(api.Project{ID: "proj_fallback", Name: "!!!"})
	if got != "proj_fallback" {
		t.Fatalf("inferProfileName fallback = %q, want proj_fallback", got)
	}
}

func TestResolveAddProjectProfileNameRequiresNameForMultipleProjects(t *testing.T) {
	projects := []api.Project{
		{ID: "proj_one", Name: "One"},
		{ID: "proj_two", Name: "Two"},
	}

	if _, err := resolveAddProjectProfileName("", projects); err == nil {
		t.Fatal("expected an error when multiple projects are returned without --name")
	}

	got, err := resolveAddProjectProfileName("explicit", projects)
	if err != nil {
		t.Fatalf("explicit name returned error: %v", err)
	}
	if got != "explicit" {
		t.Fatalf("explicit name = %q, want explicit", got)
	}
}
