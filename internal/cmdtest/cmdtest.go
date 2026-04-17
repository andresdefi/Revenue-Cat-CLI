package cmdtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"testing"

	rootcmd "github.com/andresdefi/rc/cmd"
	"github.com/andresdefi/rc/internal/api"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/config"
	"github.com/andresdefi/rc/internal/output"
)

const (
	TestProjectID = "proj_cmdtest"
	TestToken     = "sk_cmdtest_token"
	TestProfile   = "cmdtest"
)

type Request struct {
	Method string
	Path   string
	Query  string
	Body   string
}

type Result struct {
	Stdout   string
	Stderr   string
	Err      error
	Requests []Request
}

type runConfig struct {
	handler                 http.HandlerFunc
	projectID               string
	token                   string
	profile                 string
	stdin                   string
	context                 context.Context
	cancelOnRepeatedRequest context.CancelFunc
}

type Option func(*runConfig)

func WithHandler(handler http.HandlerFunc) Option {
	return func(cfg *runConfig) {
		cfg.handler = handler
	}
}

func WithAPIError(status int, typ, message string) Option {
	return WithHandler(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, status, map[string]any{
			"object":  "error",
			"type":    typ,
			"message": message,
		})
	})
}

func WithoutToken() Option {
	return func(cfg *runConfig) {
		cfg.token = ""
	}
}

func WithoutProject() Option {
	return func(cfg *runConfig) {
		cfg.projectID = ""
	}
}

func WithStdin(input string) Option {
	return func(cfg *runConfig) {
		cfg.stdin = input
	}
}

func WithContext(ctx context.Context) Option {
	return func(cfg *runConfig) {
		cfg.context = ctx
	}
}

func WithCancelOnRepeatedRequest(cancel context.CancelFunc) Option {
	return func(cfg *runConfig) {
		cfg.cancelOnRepeatedRequest = cancel
	}
}

func Run(t *testing.T, args []string, opts ...Option) Result {
	t.Helper()

	cfg := runConfig{
		handler:   DefaultHandler,
		projectID: TestProjectID,
		token:     TestToken,
		profile:   TestProfile,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	var (
		mu            sync.Mutex
		requests      []Request
		requestCounts = map[string]int{}
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		mu.Lock()
		requests = append(requests, Request{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.RawQuery,
			Body:   string(body),
		})
		key := r.Method + " " + r.URL.Path
		requestCounts[key]++
		if cfg.cancelOnRepeatedRequest != nil && requestCounts[key] == 2 {
			cfg.cancelOnRepeatedRequest()
		}
		mu.Unlock()

		if got := r.Header.Get("Authorization"); got != "Bearer "+TestToken {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"object":  "error",
				"type":    "authentication_error",
				"message": "invalid API key",
			})
			return
		}
		cfg.handler(w, r)
	}))
	defer server.Close()

	oldBaseURL := api.BaseURL
	oldDryRun := api.DryRun
	oldPrettyJSON := output.PrettyJSON
	oldForceYes := cmdutil.ForceYes
	api.BaseURL = server.URL
	api.DryRun = false
	output.PrettyJSON = true
	cmdutil.ForceYes = true
	t.Cleanup(func() {
		api.BaseURL = oldBaseURL
		api.DryRun = oldDryRun
		output.PrettyJSON = oldPrettyJSON
		cmdutil.ForceYes = oldForceYes
		cmdutil.ActiveProfile = ""
	})

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("RC_BYPASS_KEYCHAIN", "1")
	if err := writeConfig(cfg.profile, cfg.projectID, cfg.token); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cmd := rootcmd.NewRootCmd()
	if cfg.context != nil {
		cmd.SetContext(cfg.context)
	}
	cmd.SetArgs(args)
	stdout, stderr, err := capture(t, cfg.stdin, func() error {
		return cmd.Execute()
	})

	mu.Lock()
	defer mu.Unlock()
	return Result{
		Stdout:   stdout,
		Stderr:   stderr,
		Err:      err,
		Requests: append([]Request(nil), requests...),
	}
}

func AssertSuccess(t *testing.T, result Result) {
	t.Helper()
	if result.Err != nil {
		t.Fatalf("command returned error: %v\nstdout:\n%s\nstderr:\n%s", result.Err, result.Stdout, result.Stderr)
	}
}

func AssertErrorContains(t *testing.T, result Result, want string) {
	t.Helper()
	if result.Err == nil {
		t.Fatalf("expected error containing %q, got nil\nstdout:\n%s\nstderr:\n%s", want, result.Stdout, result.Stderr)
	}
	if !strings.Contains(result.Err.Error(), want) {
		t.Fatalf("error = %q, want substring %q", result.Err.Error(), want)
	}
}

func AssertOutputContains(t *testing.T, result Result, want string) {
	t.Helper()
	if !strings.Contains(result.Stdout, want) && !strings.Contains(result.Stderr, want) {
		t.Fatalf("output missing %q\nstdout:\n%s\nstderr:\n%s", want, result.Stdout, result.Stderr)
	}
}

func AssertRequested(t *testing.T, result Result, method, requestPath string) {
	t.Helper()
	for _, req := range result.Requests {
		if req.Method == method && req.Path == requestPath {
			return
		}
	}
	t.Fatalf("missing request %s %s; got %#v", method, requestPath, result.Requests)
}

func AssertRequestCountAtLeast(t *testing.T, result Result, method, requestPath string, want int) {
	t.Helper()
	got := 0
	for _, req := range result.Requests {
		if req.Method == method && req.Path == requestPath {
			got++
		}
	}
	if got < want {
		t.Fatalf("request count for %s %s = %d, want at least %d; got %#v", method, requestPath, got, want, result.Requests)
	}
}

func AssertRequestJSON(t *testing.T, result Result, method, requestPath string, want map[string]any) {
	t.Helper()
	for _, req := range result.Requests {
		if req.Method != method || req.Path != requestPath {
			continue
		}
		var got map[string]any
		if err := json.Unmarshal([]byte(req.Body), &got); err != nil {
			t.Fatalf("request body for %s %s is not JSON: %v\nbody: %s", method, requestPath, err, req.Body)
		}
		gotRaw, _ := json.Marshal(got)
		wantRaw, _ := json.Marshal(want)
		if string(gotRaw) != string(wantRaw) {
			t.Fatalf("request body for %s %s = %s, want %s", method, requestPath, gotRaw, wantRaw)
		}
		return
	}
	t.Fatalf("missing request %s %s; got %#v", method, requestPath, result.Requests)
}

func AssertRequestBody(t *testing.T, result Result, method, requestPath, want string) {
	t.Helper()
	for _, req := range result.Requests {
		if req.Method != method || req.Path != requestPath {
			continue
		}
		if req.Body != want {
			t.Fatalf("request body for %s %s = %q, want %q", method, requestPath, req.Body, want)
		}
		return
	}
	t.Fatalf("missing request %s %s; got %#v", method, requestPath, result.Requests)
}

func AssertNotRequested(t *testing.T, result Result, method, requestPath string) {
	t.Helper()
	for _, req := range result.Requests {
		if req.Method == method && req.Path == requestPath {
			t.Fatalf("unexpected request %s %s; got %#v", method, requestPath, result.Requests)
		}
	}
}

func AssertRequestedWithQuery(t *testing.T, result Result, method, requestPath, queryKey, queryValue string) {
	t.Helper()
	for _, req := range result.Requests {
		if req.Method != method || req.Path != requestPath {
			continue
		}
		values, err := url.ParseQuery(req.Query)
		if err != nil {
			t.Fatalf("request query for %s %s is invalid: %v", method, requestPath, err)
		}
		if got := values.Get(queryKey); got == queryValue {
			return
		}
	}
	t.Fatalf("missing request %s %s with %s=%s; got %#v", method, requestPath, queryKey, queryValue, result.Requests)
}

func writeConfig(profile, projectID, token string) error {
	p := &config.Profile{
		APIKey:    token,
		ProjectID: projectID,
	}
	return config.Save(&config.Config{
		CurrentProfile: profile,
		Profiles: map[string]*config.Profile{
			profile: p,
		},
	})
}

func capture(t *testing.T, stdin string, fn func() error) (string, string, error) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldStdin := os.Stdin

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	var stdinFile *os.File
	if stdin != "" {
		stdinFile, err = os.CreateTemp(t.TempDir(), "stdin-*")
		if err != nil {
			t.Fatalf("stdin temp file: %v", err)
		}
		if _, err := stdinFile.WriteString(stdin); err != nil {
			t.Fatalf("write stdin: %v", err)
		}
		if _, err := stdinFile.Seek(0, io.SeekStart); err != nil {
			t.Fatalf("seek stdin: %v", err)
		}
		os.Stdin = stdinFile
	}

	os.Stdout = stdoutW
	os.Stderr = stderrW

	stdoutCh := make(chan string)
	stderrCh := make(chan string)
	go readPipe(stdoutR, stdoutCh)
	go readPipe(stderrR, stderrCh)

	runErr := fn()

	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	os.Stdin = oldStdin
	if stdinFile != nil {
		_ = stdinFile.Close()
	}

	stdout := <-stdoutCh
	stderr := <-stderrCh
	return stdout, stderr, runErr
}

func readPipe(r *os.File, ch chan<- string) {
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	ch <- buf.String()
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGET(w, r)
	case http.MethodPost:
		handlePOST(w, r)
	case http.MethodDelete:
		writeJSON(w, http.StatusOK, map[string]any{
			"object":     "deleted",
			"id":         path.Base(r.URL.Path),
			"deleted_at": 1713072000000,
		})
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorPayload("method_not_allowed", "method not allowed"))
	}
}

func handleGET(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	switch {
	case p == "/projects":
		writeJSON(w, http.StatusOK, list(project()))
	case strings.HasSuffix(p, "/apps"):
		writeJSON(w, http.StatusOK, list(app()))
	case strings.Contains(p, "/apps/") && strings.HasSuffix(p, "/public_api_keys"):
		writeJSON(w, http.StatusOK, list(publicKey()))
	case strings.Contains(p, "/apps/") && strings.HasSuffix(p, "/store_kit_config"):
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte(`{"object":"store_kit_config","id":"skc_cmdtest"}`))
	case strings.Contains(p, "/apps/"):
		writeJSON(w, http.StatusOK, app())
	case strings.Contains(p, "/entitlements/") && strings.HasSuffix(p, "/products"):
		writeJSON(w, http.StatusOK, list(product()))
	case strings.Contains(p, "/packages/") && strings.HasSuffix(p, "/products"):
		writeJSON(w, http.StatusOK, list(packageProduct()))
	case strings.HasSuffix(p, "/products"):
		writeJSON(w, http.StatusOK, list(product()))
	case strings.Contains(p, "/products/") && strings.HasSuffix(p, "/entitlements"):
		writeJSON(w, http.StatusOK, list(entitlement()))
	case strings.Contains(p, "/products/"):
		writeJSON(w, http.StatusOK, product())
	case strings.HasSuffix(p, "/entitlements"):
		writeJSON(w, http.StatusOK, list(entitlement()))
	case strings.Contains(p, "/entitlements/"):
		writeJSON(w, http.StatusOK, entitlement())
	case strings.HasSuffix(p, "/packages"):
		writeJSON(w, http.StatusOK, list(pkg()))
	case strings.Contains(p, "/packages/"):
		writeJSON(w, http.StatusOK, pkg())
	case strings.HasSuffix(p, "/offerings"):
		writeJSON(w, http.StatusOK, list(offering()))
	case strings.Contains(p, "/offerings/"):
		writeJSON(w, http.StatusOK, offering())
	case strings.HasSuffix(p, "/customers"):
		writeJSON(w, http.StatusOK, list(customer()))
	case strings.Contains(p, "/customers/") && strings.HasSuffix(p, "/virtual_currencies"):
		writeJSON(w, http.StatusOK, list(vcBalance()))
	case strings.Contains(p, "/customers/") && strings.HasSuffix(p, "/active_entitlements"):
		writeJSON(w, http.StatusOK, list(activeEntitlement()))
	case strings.Contains(p, "/customers/") && strings.HasSuffix(p, "/subscriptions"):
		writeJSON(w, http.StatusOK, list(subscription()))
	case strings.Contains(p, "/customers/") && strings.HasSuffix(p, "/purchases"):
		writeJSON(w, http.StatusOK, list(purchase()))
	case strings.Contains(p, "/customers/") && strings.HasSuffix(p, "/aliases"):
		writeJSON(w, http.StatusOK, list(customerAlias()))
	case strings.Contains(p, "/customers/") && strings.HasSuffix(p, "/attributes"):
		writeJSON(w, http.StatusOK, list(customerAttribute()))
	case strings.Contains(p, "/customers/") && strings.HasSuffix(p, "/invoices"):
		writeJSON(w, http.StatusOK, list(invoice()))
	case strings.Contains(p, "/customers/") && strings.Contains(p, "/invoices/"):
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-1.4\n"))
	case strings.Contains(p, "/customers/"):
		writeJSON(w, http.StatusOK, customer())
	case strings.HasSuffix(p, "/subscriptions"):
		writeJSON(w, http.StatusOK, list(subscription()))
	case strings.Contains(p, "/subscriptions/") && strings.HasSuffix(p, "/transactions"):
		writeJSON(w, http.StatusOK, list(transaction()))
	case strings.Contains(p, "/subscriptions/") && strings.HasSuffix(p, "/entitlements"):
		writeJSON(w, http.StatusOK, list(activeEntitlement()))
	case strings.Contains(p, "/subscriptions/") && strings.HasSuffix(p, "/authenticated_management_url"):
		writeJSON(w, http.StatusOK, map[string]any{"object": "authenticated_management_url", "management_url": "https://pay.rev.cat/manage"})
	case strings.Contains(p, "/subscriptions/"):
		writeJSON(w, http.StatusOK, subscription())
	case strings.HasSuffix(p, "/purchases"):
		writeJSON(w, http.StatusOK, list(purchase()))
	case strings.Contains(p, "/purchases/") && strings.HasSuffix(p, "/entitlements"):
		writeJSON(w, http.StatusOK, list(activeEntitlement()))
	case strings.Contains(p, "/purchases/"):
		writeJSON(w, http.StatusOK, purchase())
	case strings.HasSuffix(p, "/webhooks"):
		writeJSON(w, http.StatusOK, list(webhook()))
	case strings.Contains(p, "/webhooks/"):
		writeJSON(w, http.StatusOK, webhook())
	case strings.HasSuffix(p, "/metrics/overview"):
		writeJSON(w, http.StatusOK, overviewMetrics())
	case strings.Contains(p, "/charts/") && strings.HasSuffix(p, "/options"):
		writeJSON(w, http.StatusOK, chartOptions())
	case strings.Contains(p, "/charts/"):
		writeJSON(w, http.StatusOK, chartData())
	case strings.HasSuffix(p, "/paywalls"):
		writeJSON(w, http.StatusOK, list(paywall()))
	case strings.Contains(p, "/paywalls/"):
		writeJSON(w, http.StatusOK, paywall())
	case strings.HasSuffix(p, "/audit_logs"):
		writeJSON(w, http.StatusOK, list(auditLog()))
	case strings.HasSuffix(p, "/collaborators"):
		writeJSON(w, http.StatusOK, list(collaborator()))
	case strings.HasSuffix(p, "/virtual_currencies"):
		writeJSON(w, http.StatusOK, list(virtualCurrency()))
	case strings.Contains(p, "/virtual_currencies/"):
		writeJSON(w, http.StatusOK, virtualCurrency())
	default:
		writeJSON(w, http.StatusNotFound, errorPayload("not_found", fmt.Sprintf("no fixture for GET %s", p)))
	}
}

func handlePOST(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	switch {
	case p == "/projects":
		writeJSON(w, http.StatusCreated, project())
	case strings.HasSuffix(p, "/apps"):
		writeJSON(w, http.StatusCreated, app())
	case strings.Contains(p, "/apps/"):
		writeJSON(w, http.StatusOK, app())
	case strings.HasSuffix(p, "/products"):
		writeJSON(w, http.StatusCreated, product())
	case strings.Contains(p, "/products/"):
		writeJSON(w, http.StatusOK, product())
	case strings.HasSuffix(p, "/entitlements"):
		writeJSON(w, http.StatusCreated, entitlement())
	case strings.Contains(p, "/entitlements/"):
		writeJSON(w, http.StatusOK, entitlement())
	case strings.HasSuffix(p, "/packages"):
		writeJSON(w, http.StatusCreated, pkg())
	case strings.Contains(p, "/packages/"):
		writeJSON(w, http.StatusOK, pkg())
	case strings.HasSuffix(p, "/offerings"):
		writeJSON(w, http.StatusCreated, offering())
	case strings.Contains(p, "/offerings/"):
		writeJSON(w, http.StatusOK, offering())
	case strings.HasSuffix(p, "/customers"):
		writeJSON(w, http.StatusCreated, customer())
	case strings.Contains(p, "/customers/") && strings.HasSuffix(p, "/virtual_currencies/transactions"):
		writeJSON(w, http.StatusOK, vcTransaction())
	case strings.Contains(p, "/customers/") && strings.HasSuffix(p, "/virtual_currencies/update_balance"):
		writeJSON(w, http.StatusOK, vcBalance())
	case strings.Contains(p, "/customers/"):
		writeJSON(w, http.StatusOK, customer())
	case strings.Contains(p, "/subscriptions/") && strings.HasSuffix(p, "/authenticated_management_url"):
		writeJSON(w, http.StatusOK, map[string]any{"object": "authenticated_management_url", "management_url": "https://pay.rev.cat/manage"})
	case strings.Contains(p, "/subscriptions/"):
		writeJSON(w, http.StatusOK, subscription())
	case strings.Contains(p, "/purchases/"):
		writeJSON(w, http.StatusOK, purchase())
	case strings.HasSuffix(p, "/webhooks"):
		writeJSON(w, http.StatusCreated, webhook())
	case strings.Contains(p, "/webhooks/"):
		writeJSON(w, http.StatusOK, webhook())
	case strings.HasSuffix(p, "/paywalls"):
		writeJSON(w, http.StatusCreated, paywall())
	case strings.Contains(p, "/paywalls/"):
		writeJSON(w, http.StatusOK, paywall())
	case strings.HasSuffix(p, "/virtual_currencies"):
		writeJSON(w, http.StatusCreated, virtualCurrency())
	case strings.Contains(p, "/virtual_currencies/"):
		writeJSON(w, http.StatusOK, virtualCurrency())
	default:
		writeJSON(w, http.StatusNotFound, errorPayload("not_found", fmt.Sprintf("no fixture for POST %s", p)))
	}
}

func list(items ...any) map[string]any {
	return map[string]any{
		"object":    "list",
		"items":     items,
		"next_page": nil,
		"url":       "/fixture",
	}
}

func errorPayload(typ, message string) map[string]any {
	return map[string]any{
		"object":  "error",
		"type":    typ,
		"message": message,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func project() map[string]any {
	return map[string]any{"object": "project", "id": TestProjectID, "name": "Command Test Project", "created_at": 1713072000000}
}

func app() map[string]any {
	return map[string]any{
		"object":     "app",
		"id":         "app_cmdtest",
		"name":       "iOS App",
		"type":       "app_store",
		"project_id": TestProjectID,
		"created_at": 1713072000000,
		"app_store": map[string]any{
			"bundle_id":                            "com.example.app",
			"app_store_connect_api_key_configured": true,
			"subscription_key_configured":          true,
		},
	}
}

func publicKey() map[string]any {
	return map[string]any{"object": "public_api_key", "key": "appl_public_key", "name": "iOS public key"}
}

func product() map[string]any {
	return map[string]any{"object": "product", "id": "prod_cmdtest", "store_identifier": "com.example.premium.monthly", "type": "subscription", "state": "active", "display_name": "Premium Monthly", "app_id": "app_cmdtest", "created_at": 1713072000000}
}

func entitlement() map[string]any {
	return map[string]any{"object": "entitlement", "id": "entl_cmdtest", "project_id": TestProjectID, "lookup_key": "premium", "display_name": "Premium", "state": "active", "created_at": 1713072000000}
}

func offering() map[string]any {
	return map[string]any{"object": "offering", "id": "ofrnge_cmdtest", "project_id": TestProjectID, "lookup_key": "default", "display_name": "Default", "is_current": true, "state": "active", "created_at": 1713072000000}
}

func pkg() map[string]any {
	position := 1
	return map[string]any{"object": "package", "id": "pkge_cmdtest", "lookup_key": "$rc_monthly", "display_name": "Monthly", "position": position, "created_at": 1713072000000}
}

func packageProduct() map[string]any {
	return map[string]any{"object": "package_product", "product_id": "prod_cmdtest", "eligibility_criteria": "all", "product": product()}
}

func customer() map[string]any {
	return map[string]any{"object": "customer", "id": "cust_cmdtest", "project_id": TestProjectID, "first_seen_at": 1713072000000, "active_entitlements": list(activeEntitlement())}
}

func activeEntitlement() map[string]any {
	return map[string]any{"object": "active_entitlement", "entitlement_id": "entl_cmdtest", "expires_at": nil}
}

func customerAlias() map[string]any {
	return map[string]any{"object": "customer_alias", "id": "alias_cmdtest"}
}

func customerAttribute() map[string]any {
	return map[string]any{"object": "customer_attribute", "name": "email", "value": "customer@example.com"}
}

func subscription() map[string]any {
	return map[string]any{"object": "subscription", "id": "sub_cmdtest", "customer_id": "cust_cmdtest", "original_customer_id": "cust_cmdtest", "product_id": "prod_cmdtest", "starts_at": 1713072000000, "current_period_starts_at": 1713072000000, "current_period_ends_at": 1715750400000, "ends_at": nil, "gives_access": true, "pending_payment": false, "auto_renewal_status": "will_renew", "status": "active", "presented_offering_id": "ofrnge_cmdtest", "environment": "production", "store": "app_store", "store_subscription_identifier": "store_sub_cmdtest", "ownership": "purchased", "country": "US", "management_url": "https://pay.rev.cat/manage"}
}

func transaction() map[string]any {
	return map[string]any{"object": "transaction", "id": "txn_cmdtest", "purchased_at": 1713072000000, "store": "app_store", "revenue_in_usd": money()}
}

func money() map[string]any {
	return map[string]any{"currency": "USD", "gross": 9.99, "commission": 1.49, "tax": 0.0, "proceeds": 8.50}
}

func purchase() map[string]any {
	return map[string]any{"object": "purchase", "id": "purch_cmdtest", "customer_id": "cust_cmdtest", "original_customer_id": "cust_cmdtest", "product_id": "prod_cmdtest", "purchased_at": 1713072000000, "revenue_in_usd": money(), "quantity": 1, "status": "active", "presented_offering_id": "ofrnge_cmdtest", "environment": "production", "store": "app_store", "store_purchase_identifier": "store_purchase_cmdtest", "ownership": "purchased", "country": "US"}
}

func invoice() map[string]any {
	return map[string]any{"object": "invoice", "id": "inv_cmdtest", "created_at": 1713072000000}
}

func webhook() map[string]any {
	return map[string]any{"object": "webhook", "id": "wh_cmdtest", "name": "Events", "url": "https://example.com/revenuecat", "created_at": 1713072000000}
}

func overviewMetrics() map[string]any {
	return map[string]any{"object": "overview_metrics", "metrics": []any{map[string]any{"object": "metric_summary", "name": "revenue", "value": 1234.56, "description": "Revenue", "period": "last_30_days", "updated_at": 1713072000000}}}
}

func chartData() map[string]any {
	return map[string]any{"object": "chart_data", "name": "revenue", "display_name": "Revenue", "values": []any{map[string]any{"date": "2026-04-14", "value": 123.45}}}
}

func chartOptions() map[string]any {
	return map[string]any{"object": "chart_options", "options": []any{map[string]any{"name": "country", "values": []string{"US", "NO"}}}}
}

func paywall() map[string]any {
	return map[string]any{"object": "paywall", "id": "paywall_cmdtest", "created_at": 1713072000000}
}

func auditLog() map[string]any {
	return map[string]any{"object": "audit_log_entry", "id": "audit_cmdtest", "action": "product.created", "actor": "team@example.com", "details": "Created product", "created_at": 1713072000000}
}

func collaborator() map[string]any {
	return map[string]any{"object": "collaborator", "id": "collab_cmdtest", "email": "team@example.com", "role": "admin"}
}

func virtualCurrency() map[string]any {
	return map[string]any{"object": "virtual_currency", "code": "COIN", "name": "Coins", "state": "active", "created_at": 1713072000000}
}

func vcBalance() map[string]any {
	return map[string]any{"object": "virtual_currency_balance", "currency_code": "COIN", "balance": 100}
}

func vcTransaction() map[string]any {
	return map[string]any{"object": "virtual_currency_transaction", "id": "vctxn_cmdtest", "currency_code": "COIN", "amount": 100, "created_at": 1713072000000}
}
