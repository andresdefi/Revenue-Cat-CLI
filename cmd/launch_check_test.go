package cmd_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestLaunchCheckJSONReady(t *testing.T) {
	result := cmdtest.Run(t, []string{"launch-check", "--output", "json"})

	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"object": "launch_check_report"`)
	cmdtest.AssertOutputContains(t, result, `"ready": true`)
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/apps")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/packages/pkge_cmdtest/products")
}

func TestLaunchCheckTableReady(t *testing.T) {
	result := cmdtest.Run(t, []string{"launch-check", "--output", "table"})

	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "READY: YES")
	cmdtest.AssertOutputContains(t, result, "One current active offering is configured")
}

func TestLaunchCheckReportsNotReady(t *testing.T) {
	result := cmdtest.Run(t, []string{"launch-check", "--output", "json"}, cmdtest.WithHandler(launchCheckNotReadyHandler))

	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"ready": false`)
	cmdtest.AssertOutputContains(t, result, "At least one app is required")
	cmdtest.AssertOutputContains(t, result, "Exactly one current active offering is required")
}

func TestLaunchCheckStrictFailsWhenNotReady(t *testing.T) {
	result := cmdtest.Run(t, []string{"launch-check", "--strict"}, cmdtest.WithHandler(launchCheckNotReadyHandler))

	cmdtest.AssertErrorContains(t, result, "launch check failed")
}

func launchCheckNotReadyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeLaunchCheckJSON(w, http.StatusMethodNotAllowed, map[string]any{"object": "error", "type": "method_not_allowed", "message": "method not allowed"})
		return
	}

	switch p := r.URL.Path; {
	case strings.HasSuffix(p, "/apps"):
		writeLaunchCheckJSON(w, http.StatusOK, launchCheckList())
	case strings.HasSuffix(p, "/products") && !strings.Contains(p, "/entitlements/") && !strings.Contains(p, "/packages/"):
		writeLaunchCheckJSON(w, http.StatusOK, launchCheckList())
	case strings.HasSuffix(p, "/entitlements"):
		writeLaunchCheckJSON(w, http.StatusOK, launchCheckList())
	case strings.HasSuffix(p, "/offerings"):
		writeLaunchCheckJSON(w, http.StatusOK, launchCheckList())
	default:
		writeLaunchCheckJSON(w, http.StatusOK, launchCheckList())
	}
}

func launchCheckList(items ...any) map[string]any {
	return map[string]any{"object": "list", "items": items, "next_page": nil}
}

func writeLaunchCheckJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
