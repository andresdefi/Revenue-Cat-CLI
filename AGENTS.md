# rc - RevenueCat CLI

## Project Overview
Unofficial open-source CLI for the RevenueCat REST API v2. Written in Go with Cobra.
Repo: `andresdefi/Revenue-Cat-CLI` on GitHub. Binary name: `rc`.
100% API v2 coverage - 130+ command reference entries covering all 95 API endpoints across 16 command groups, plus first-run and diagnostic commands.

## Tech Stack
- **Language:** Go 1.25+ (go.mod directive: 1.25.0)
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
  foundation.go                  init, doctor, whoami
  auth/auth.go                   auth login/status/logout/doctor/validate
  config/config.go               config profiles
  projects/projects.go           projects list/create/doctor/set-default
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
  paywalls/paywalls.go           paywalls list/get/create/delete
  auditlogs/auditlogs.go        audit-logs list
  collaborators/collaborators.go collaborators list
  currencies/currencies.go       currencies list/get/create/update/delete/archive/unarchive/balance/credit/set-balance
  mcp/mcp.go                     mcp serve
  transfer/transfer.go           export/import project configuration
internal/
  api/client.go                  HTTP client with retry, auth header, error handling
  api/types.go                   All RevenueCat API v2 response types
  auth/auth.go                   Token storage (keychain + config fallback)
  config/config.go               TOML config file (~/.rc/config.toml) + legacy JSON migration
  cmdutil/cmdutil.go             Shared helpers (ResolveProject, GetOutputFormat)
  cmdtest/cmdtest.go             Command test harness + request/pagination assertions
  commanddocs/commanddocs.go     Generated Cobra command reference
  output/output.go               JSON + table + markdown output formatting
  projecthealth/projecthealth.go Read-only project setup analyzer for workflow commands
docs/
  COMMANDS.md                    Generated command reference
  WORKFLOWS.md                   Copyable setup, offering, customer, and migration recipes
  API_NOTES.md                   RevenueCat semantics and import/export notes
  CI_CD.md                       Automation examples and auth guidance
  TESTING.md                     Local, fixture, pagination, and integration test guidance
```

## Key Patterns
- **Auth:** Bearer token via `Authorization: Bearer <sk_...>`. Stored in system keychain, falls back to profile entries in `~/.rc/config.toml`
- **Profiles:** `--profile` flag > `RC_PROFILE` env var > `current_profile` > `default`
- **Project resolution:** `--project` flag > active profile default project > error with guidance
- **Output:** `--output table`, `--output json`, or `--output markdown`. Table uses go-pretty, JSON uses encoding/json
- **API client:** All requests go through `internal/api/client.go`. Retries on `retryable: true` errors with backoff
- **Error handling:** API errors parsed into `api.Error` struct with type, message, doc_url
- **Pagination:** Cursor-based list pagination is implemented with `--all`. `next_page` can be an absolute URL, a `/v2/...` path, or a bare resource path
- **No import cycles:** Command packages import `internal/cmdutil`, not `cmd`
- **Archive pattern:** Products, entitlements, offerings, and virtual currencies all support archive/unarchive
- **Attach/detach pattern:** Products can be attached/detached from both entitlements and packages
- **Project transfer:** `rc export`/`rc import` is beta. It carries apps, products, entitlements, offerings, packages, attachments, metadata, and archive/current state where the API allows it
- **Project health:** `rc project doctor` reads apps, products, entitlements, offerings, packages, and package products to report setup issues. `--strict` returns non-zero on failed health checks
- **Generated docs:** `docs/COMMANDS.md` is generated from Cobra. Run `make docs` after command changes
- **Correctness harness:** Request-body golden tests, pagination contract tests, and opt-in integration tests guard API semantics

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
make docs           # Regenerate docs/COMMANDS.md
make check          # Run all checks (fmt, docs, vet, lint, test)
make help           # Show all targets
```

## Full API Coverage (99 subcommands, 95 API endpoints)
- [x] Auth: login, status, logout
- [x] Foundation: init, doctor, whoami, config profiles, auth validate
- [x] Projects: list, create, doctor, set-default
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
- [x] JSON + table + markdown output (TTY-aware: table for terminal, JSON for pipes)
- [x] Version command with ldflags injection (rc version)
- [x] "Did you mean?" fuzzy command suggestions
- [x] Structured exit codes (1=general, 3=auth, 4=API)
- [x] 546 default tests across 37 test files (550 with integration tag)
- [x] Request golden tests, pagination contract tests, generated docs drift test, opt-in integration tests
- [x] Makefile with build/test/lint/fmt/docs/check targets
- [x] .golangci.yml config
- [x] README with badges, TOC, workflows, troubleshooting, docs index
- [x] Community files: SECURITY.md, CODE_OF_CONDUCT.md, SUPPORT.md
- [x] GitHub: issue/PR templates, topics, Discussions enabled
- [x] CI: build (Go 1.25 + stable), lint, CodeQL, govulncheck, gosec
- [x] GoReleaser config with Homebrew tap + ldflags
- [x] Install script (curl | sh)
- [x] Pre-commit hook

## Release
- **v0.2.0** released 2026-04-15
- Homebrew: `brew install andresdefi/tap/rc`
- Install script: `curl -fsSL https://raw.githubusercontent.com/andresdefi/Revenue-Cat-CLI/main/install.sh | sh`
- Go: `go install github.com/andresdefi/rc@latest`

## Future Improvements
- [ ] Interactive mode for create commands (prompt for required fields)
- [ ] `--watch` flag for polling commands
- [ ] More edge-case tests for project transfer and import/export fixtures
- [ ] Documentation website
- [ ] Apple code signing for macOS binaries
- [ ] Set up HOMEBREW_TAP_TOKEN secret for auto formula updates on release

## Module Path
`github.com/andresdefi/rc`

When the GitHub repo name is `Revenue-Cat-CLI`, the module path stays `github.com/andresdefi/rc` for clean imports. The `go install` command uses the module path.
