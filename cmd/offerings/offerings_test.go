package offerings_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/andresdefi/rc/internal/cmdtest"
)

func TestOfferingsListTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "ofrnge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings")
}

func TestOfferingsListJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"object\": \"list\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings")
}

func TestOfferingsListWithProfile(t *testing.T) {
	result := cmdtest.Run(t, []string{"--profile", "cmdtest", "offerings", "list", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "ofrnge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings")
}

func TestOfferingsListProjectFlagOverridesDefault(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list", "--project", "proj_override", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "ofrnge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_override/offerings")
}

func TestOfferingsListNotLoggedIn(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list"}, cmdtest.WithoutToken())
	cmdtest.AssertErrorContains(t, result, "not logged in")
}

func TestOfferingsListAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestOfferingsGetTable(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "get", "ofrnge_cmdtest", "--output", "table"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "ofrnge_cmdtest")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest")
}

func TestOfferingsGetJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "get", "ofrnge_cmdtest", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "\"id\": \"ofrnge_cmdtest\"")
	cmdtest.AssertRequested(t, result, "GET", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest")
}

func TestOfferingsGetMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "get"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestOfferingsCreateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "create", "--lookup-key", "default", "--display-name", "Default", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "ofrnge_cmdtest")
	cmdtest.AssertOutputContains(t, result, "next: rc packages create --offering-id ofrnge_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/offerings")
}

func TestOfferingsCreateMissingRequiredFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "create"})
	cmdtest.AssertErrorContains(t, result, "required")
}

func TestOfferingsDeleteSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "delete", "ofrnge_cmdtest"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "deleted")
	cmdtest.AssertRequested(t, result, "DELETE", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest")
}

func TestOfferingsDeleteMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "delete"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestOfferingsInvalidOutputFlag(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list", "--output", "yaml"})
	cmdtest.AssertErrorContains(t, result, "invalid output format")
}

func TestOfferingsHelpExamples(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "list", "--help"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "Examples:")
}

func TestOfferingsUpdateJSON(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "update", "ofrnge_cmdtest", "--display-name", "Default Plus", "--output", "json"})
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, "ofrnge_cmdtest")
	cmdtest.AssertRequested(t, result, "POST", "/projects/proj_cmdtest/offerings/ofrnge_cmdtest")
}

func TestOfferingsUpdateMissingArg(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "update"})
	cmdtest.AssertErrorContains(t, result, "accepts 1 arg")
}

func TestOfferingsUpdateAPIError(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "update", "ofrnge_cmdtest", "--display-name", "Default Plus", "--output", "json"}, cmdtest.WithAPIError(400, "parameter_error", "fixture API error"))
	cmdtest.AssertErrorContains(t, result, "fixture API error")
}

func TestOfferingsPublishSuccess(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "publish", "ofrnge_publish", "--output", "json"}, cmdtest.WithHandler(offeringPublishHandler(false, true)))
	cmdtest.AssertSuccess(t, result)
	cmdtest.AssertOutputContains(t, result, `"published": true`)
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_cmdtest/offerings/ofrnge_publish")
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_cmdtest/offerings/ofrnge_publish/packages")
	cmdtest.AssertRequested(t, result, http.MethodGet, "/projects/proj_cmdtest/packages/pkge_publish/products")
	cmdtest.AssertRequested(t, result, http.MethodPost, "/projects/proj_cmdtest/offerings/ofrnge_publish")
}

func TestOfferingsPublishFailsWhenPackageHasNoProducts(t *testing.T) {
	result := cmdtest.Run(t, []string{"offerings", "publish", "ofrnge_publish", "--output", "table"}, cmdtest.WithHandler(offeringPublishHandler(false, false)))
	cmdtest.AssertErrorContains(t, result, "offering publish validation failed")
	cmdtest.AssertOutputContains(t, result, "Some packages have no product links.")
}

func offeringPublishHandler(current, hasProducts bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(p, "/offerings/ofrnge_publish"):
			writeOfferingJSON(w, map[string]any{"object": "offering", "id": "ofrnge_publish", "project_id": cmdtest.TestProjectID, "lookup_key": "default", "display_name": "Default", "is_current": current, "state": "active", "created_at": 1713072000000})
		case r.Method == http.MethodGet && strings.HasSuffix(p, "/offerings/ofrnge_publish/packages"):
			writeOfferingJSON(w, offeringList(map[string]any{"object": "package", "id": "pkge_publish", "lookup_key": "$rc_monthly", "display_name": "Monthly", "created_at": 1713072000000}))
		case r.Method == http.MethodGet && strings.HasSuffix(p, "/packages/pkge_publish/products") && hasProducts:
			writeOfferingJSON(w, offeringList(map[string]any{"object": "package_product", "product_id": "prod_cmdtest", "eligibility_criteria": "all"}))
		case r.Method == http.MethodGet && strings.HasSuffix(p, "/packages/pkge_publish/products"):
			writeOfferingJSON(w, offeringList())
		case r.Method == http.MethodPost && strings.HasSuffix(p, "/offerings/ofrnge_publish"):
			writeOfferingJSON(w, map[string]any{"object": "offering", "id": "ofrnge_publish", "project_id": cmdtest.TestProjectID, "lookup_key": "default", "display_name": "Default", "is_current": true, "state": "active", "created_at": 1713072000000})
		default:
			writeOfferingJSON(w, map[string]any{"object": "error", "type": "not_found", "message": "not found"})
		}
	}
}

func offeringList(items ...any) map[string]any {
	return map[string]any{"object": "list", "items": items, "next_page": nil, "url": "/fixture"}
}

func writeOfferingJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
