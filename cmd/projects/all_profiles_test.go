package projects

import (
	"fmt"
	"net/url"
	"testing"

	internalAuth "github.com/andresdefi/rc/internal/auth"
)

type fakeProjectClient struct {
	data []byte
	err  error
}

func (c fakeProjectClient) Get(path string, query url.Values) ([]byte, error) {
	if path != "/projects" {
		return nil, fmt.Errorf("unexpected path %s", path)
	}
	return c.data, c.err
}

func TestFetchProjectsByProfileAggregatesWithProfile(t *testing.T) {
	profiles := []internalAuth.StoredProfile{
		{Name: "default", Token: "sk_default"},
		{Name: "impostor", Token: "sk_impostor"},
	}
	clients := map[string]fakeProjectClient{
		"sk_default": {
			data: []byte(`{"object":"list","items":[{"object":"project","id":"proj_default","name":"Spentio","created_at":1776067200000}],"next_page":null}`),
		},
		"sk_impostor": {
			data: []byte(`{"object":"list","items":[{"object":"project","id":"proj_impostor","name":"Impostor","created_at":1776240000000}],"next_page":null}`),
		},
	}

	items, warnings := fetchProjectsByProfile(profiles, nil, func(token string) projectListClient {
		return clients[token]
	})

	if len(warnings) != 0 {
		t.Fatalf("warnings = %#v, want none", warnings)
	}
	if len(items) != 2 {
		t.Fatalf("items len = %d, want 2", len(items))
	}
	if items[0].Profile != "default" || items[0].ID != "proj_default" {
		t.Fatalf("first item = %#v, want default/proj_default", items[0])
	}
	if items[1].Profile != "impostor" || items[1].ID != "proj_impostor" {
		t.Fatalf("second item = %#v, want impostor/proj_impostor", items[1])
	}
}

func TestFetchProjectsByProfileSkipsFailedProfile(t *testing.T) {
	profiles := []internalAuth.StoredProfile{
		{Name: "default", Token: "sk_default"},
		{Name: "expired", Token: "sk_expired"},
	}
	clients := map[string]fakeProjectClient{
		"sk_default": {
			data: []byte(`{"object":"list","items":[{"object":"project","id":"proj_default","name":"Spentio","created_at":1776067200000}],"next_page":null}`),
		},
		"sk_expired": {
			err: fmt.Errorf("authentication_error: invalid API key"),
		},
	}

	items, warnings := fetchProjectsByProfile(profiles, nil, func(token string) projectListClient {
		return clients[token]
	})

	if len(items) != 1 || items[0].Profile != "default" {
		t.Fatalf("items = %#v, want one default item", items)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings len = %d, want 1", len(warnings))
	}
	if warnings[0].Profile != "expired" {
		t.Fatalf("warning profile = %q, want expired", warnings[0].Profile)
	}
}
