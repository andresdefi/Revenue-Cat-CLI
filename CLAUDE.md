# rc - RevenueCat CLI

## Project Overview
Unofficial open-source CLI for the RevenueCat REST API v2. Written in Go with Cobra.
Repo: `andresdefi/Revenue-Cat-CLI` on GitHub. Binary name: `rc`.
100% API v2 coverage - 99 subcommands covering all 95 API endpoints across 16 command groups.

## Tech Stack
- **Language:** Go 1.23+ (go.mod directive: 1.23)
- **CLI framework:** Cobra (github.com/spf13/cobra)
- **Table output:** go-pretty/v6 (github.com/jedib0t/go-pretty/v6)
- **Keychain:** go-keyring (github.com/zalando/go-keyring)
- **Release:** GoReleaser, Homebrew tap
- **CI:** GitHub Actions (build, test, lint, release)

## Architecture
```
main.go                          Entry point
cmd/
  root.go                        Root command, persistent flags (--project, --output)
  auth/auth.go                   auth login/status/logout
  projects/projects.go           projects list/create/set-default
  apps/apps.go                   apps list/get/create/update/delete
  products/products.go           products list/get/create/update/delete/archive/unarchive
  entitlements/entitlements.go   entitlements list/get/create/update/delete/archive/unarchive/products/attach/detach
  offerings/offerings.go         offerings list/get/create/update/delete/archive/unarchive
  packages/packages.go           packages list/get/create/update/delete/products/attach/detach
  customers/customers.go         customers list/lookup/create/delete/entitlements/subscriptions/purchases/aliases/attributes/set-attributes/grant/revoke/assign-offering/transfer
  subscriptions/subscriptions.go subscriptions list/get/transactions/entitlements/cancel/refund/refund-transaction/management-url
  purchases/purchases.go         purchases list/get/entitlements/refund
  webhooks/webhooks.go           webhooks list/get/create/update/delete
  charts/charts.go               charts overview/show
  paywalls/paywalls.go           paywalls list/get/delete
  auditlogs/auditlogs.go        audit-logs list
  collaborators/collaborators.go collaborators list
  currencies/currencies.go       currencies list/get/create/update/delete/archive/unarchive/balance/credit/set-balance
internal/
  api/client.go                  HTTP client with retry, auth header, error handling
  api/types.go                   All RevenueCat API v2 response types
  auth/auth.go                   Token storage (keychain + config fallback)
  config/config.go               Config file (~/.rc/config.json)
  cmdutil/cmdutil.go             Shared helpers (ResolveProject, GetOutputFormat)
  output/output.go               JSON + table output formatting
```

## Key Patterns
- **Auth:** Bearer token via `Authorization: Bearer <sk_...>`. Stored in system keychain, falls back to `~/.rc/config.json`
- **Project resolution:** `--project` flag > config default > error with guidance
- **Output:** `--output table` (default) or `--output json`. Table uses go-pretty, JSON uses encoding/json
- **API client:** All requests go through `internal/api/client.go`. Retries on `retryable: true` errors with backoff
- **Error handling:** API errors parsed into `api.Error` struct with type, message, doc_url
- **Pagination:** Cursor-based with `starting_after` param (not yet implemented - lists return first page)
- **No import cycles:** Command packages import `internal/cmdutil`, not `cmd`
- **Archive pattern:** Products, entitlements, offerings, and virtual currencies all support archive/unarchive
- **Attach/detach pattern:** Products can be attached/detached from both entitlements and packages

## RevenueCat API v2 Reference
- **Base URL:** `https://api.revenuecat.com/v2`
- **Auth:** `Authorization: Bearer <token>` (sk_ secret keys or atk_ OAuth tokens)
- **All updates use POST** (not PUT/PATCH)
- **Timestamps:** int64 milliseconds since epoch
- **List envelope:** `{ "object": "list", "items": [...], "next_page": "..." }`
- **Error envelope:** `{ "object": "error", "type": "...", "message": "...", "doc_url": "..." }`
- **ID prefixes:** proj, app, prod, entl, ofrnge, pkge, sub, purch
- **Rate limits:** Project config 60/min, Customer info 480/min, Charts 5/min

## Build & Run
```bash
make build          # Build with version injection
make test           # Run tests
make lint           # Run linter
make check          # Run all checks (fmt, vet, lint, test)
make help           # Show all targets
```

## Full API Coverage (99 subcommands, 95 API endpoints)
- [x] Auth: login, status, logout
- [x] Projects: list, create, set-default
- [x] Apps: list, get, create, update, delete, public-keys, storekit-config
- [x] Products: list, get, create, update, delete, archive, unarchive, push-to-store
- [x] Entitlements: list, get, create, update, delete, archive, unarchive, products, attach, detach
- [x] Offerings: list, get, create, update, delete, archive, unarchive
- [x] Packages: list, get, create, update, delete, products, attach, detach
- [x] Customers: list, lookup, create, delete, entitlements, subscriptions, purchases, aliases, attributes, set-attributes, grant, revoke, assign-offering, transfer, restore-purchase, invoices, invoice-file
- [x] Subscriptions: list, get, transactions, entitlements, cancel, refund, refund-transaction, management-url
- [x] Purchases: list, get, entitlements, refund
- [x] Webhooks: list, get, create, update, delete
- [x] Charts: overview, show, options
- [x] Paywalls: list, get, create, delete
- [x] Audit Logs: list
- [x] Collaborators: list
- [x] Virtual Currencies: list, get, create, update, delete, archive, unarchive, balance, credit, set-balance
- [x] JSON + table output (TTY-aware: table for terminal, JSON for pipes)
- [x] Version command with ldflags injection (rc version)
- [x] "Did you mean?" fuzzy command suggestions
- [x] Structured exit codes (1=general, 3=auth, 4=API)
- [x] JSON + table output (TTY-aware: table for terminal, JSON for pipes)
- [x] MCP server (rc mcp serve) - 16 tools via official Go MCP SDK
- [x] Pagination: --all and --limit on all 14 list commands
- [x] Multi-profile auth: --profile flag, RC_PROFILE env, TOML config with migration
- [x] Watch mode: --watch on customers lookup/entitlements, subscriptions get, charts overview
- [x] Bulk import/export: products and entitlements (CSV + JSON)
- [x] Interactive mode: products/entitlements/offerings create prompt when TTY
- [x] Project config transfer: rc export / rc import
- [x] 522 tests, including command-level Cobra integration tests with httptest-backed RevenueCat API fixtures
- [x] Makefile, .golangci.yml, pre-commit hook
- [x] README with badges, TOC, workflows, troubleshooting
- [x] Community files: SECURITY.md, CODE_OF_CONDUCT.md, SUPPORT.md
- [x] GitHub: issue/PR templates, topics, Discussions enabled
- [x] CI: build (Go 1.25 + stable), lint (golangci-lint v2), CodeQL, govulncheck
- [x] GoReleaser with Homebrew tap + ldflags
- [x] Install script (curl | sh)

## Release
- **v0.1.0** released 2026-04-13 - initial release, 100% API coverage
- **v0.2.0** - MCP server, pagination, multi-profile, watch, bulk ops, interactive mode
- Homebrew: `brew install andresdefi/tap/rc`
- Install script: `curl -fsSL https://raw.githubusercontent.com/andresdefi/Revenue-Cat-CLI/main/install.sh | sh`
- Go: `go install github.com/andresdefi/rc@latest`

## New Dependencies (v0.2.0)
- `github.com/modelcontextprotocol/go-sdk` - MCP server
- `github.com/charmbracelet/huh` - interactive prompts
- `github.com/jszwec/csvutil` - CSV marshal/unmarshal
- `github.com/BurntSushi/toml` - TOML config with profiles

## Future Improvements
- [x] Rich workflow-oriented --help examples for high-traffic auth, products, entitlements, offerings, customers, and subscriptions commands
- [ ] Command grouping in help output (by category)
- [x] More tests (command integration tests with mock API server)
- [ ] Documentation website
- [x] Apple Developer ID signing and notarization hooks for macOS binaries (skips gracefully without secrets)
- [ ] Set up HOMEBREW_TAP_TOKEN secret for auto formula updates on release

## Module Path
`github.com/andresdefi/rc`

When the GitHub repo name is `Revenue-Cat-CLI`, the module path stays `github.com/andresdefi/rc` for clean imports. The `go install` command uses the module path.
