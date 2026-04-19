# rc - RevenueCat CLI

## Project Overview
Unofficial open-source CLI for the RevenueCat REST API v2. Written in Go with Cobra.
Repo: `andresdefi/Revenue-Cat-CLI` on GitHub. Binary name: `rc`.
100% API v2 coverage - 104 subcommands covering all 95 API endpoints across 16 command groups.

## Tech Stack
- **Language:** Go 1.25+ (go.mod directive: 1.25)
- **CLI framework:** Cobra (github.com/spf13/cobra)
- **Table output:** go-pretty/v6 (github.com/jedib0t/go-pretty/v6)
- **Keychain:** go-keyring (github.com/zalando/go-keyring)
- **Interactive prompts:** charmbracelet/huh
- **Release:** GoReleaser (with macOS signing/notarization), Homebrew tap
- **CI:** GitHub Actions (build, test, lint, gosec, govulncheck, CodeQL, release)

## Architecture
```
main.go                          Entry point
cmd/
  root.go                        Root command, persistent flags, update check, exit codes
  completion.go                  Shell completion (bash/zsh/fish/powershell)
  auth/auth.go                   auth login/status/logout/doctor
  projects/projects.go           projects list/create/set-default (with dynamic completions)
  setup/                         setup product workflow
  apps/                          apps CRUD (with dynamic completions)
  products/                      products CRUD + archive + bulk import/export
  entitlements/                  entitlements CRUD + archive + attach/detach + bulk
  offerings/                     offerings CRUD + publish + archive
  packages/                      packages CRUD + attach/detach
  customers/                     customers CRUD + entitlements/subscriptions/purchases/grant/revoke
  subscriptions/                 subscriptions CRUD + cancel/refund
  purchases/                     purchases CRUD + refund
  webhooks/                      webhooks CRUD
  charts/                        charts overview/show/options
  paywalls/                      paywalls CRUD + validate
  auditlogs/                     audit-logs list
  collaborators/                 collaborators list
  currencies/                    virtual currencies CRUD + balance/credit
  mcp/                           MCP server (16 tools)
  transfer/                      [beta] project config export/import + migrate dry-run
internal/
  api/client.go                  HTTP client: retry, Retry-After, verbose logging, dry-run, caching
  api/types.go                   All RevenueCat API v2 response types
  auth/auth.go                   Token storage (keychain + config fallback, RC_BYPASS_KEYCHAIN)
  cache/cache.go                 File-based response cache (~/.rc/cache/) with TTL
  cmdutil/cmdutil.go             Shared helpers (ResolveProject, GetOutputFormat, ForceYes)
  cmdutil/interactive.go         Confirmation prompts (PromptConfirm, ConfirmDestructive)
  cmdutil/stability.go           Stability labels ([experimental], [beta])
  completions/completions.go     Dynamic shell completions for resource IDs (8 types)
  config/config.go               TOML config with validation and legacy JSON migration
  exitcode/exitcode.go           Granular exit codes: 4xx->10+(s-400), 5xx->60+(s-500)
  output/output.go               JSON/table/markdown output, field selection, log levels, color
  update/update.go               Non-blocking update check against GitHub releases
  validate/validate.go           Input validation helpers
```

## Key Patterns
- **Auth:** Bearer token via `Authorization: Bearer <sk_...>`. Stored in system keychain, falls back to `~/.rc/config.toml`. Strict prefix validation (sk_/atk_)
- **Profile resolution:** `--profile` flag > `RC_PROFILE` env > config `current_profile` > "default"
- **Project resolution:** `--project` flag > config default > error with guidance
- **Output:** `--output table|json|markdown`. Table uses go-pretty, JSON uses encoding/json, Markdown uses go-pretty RenderMarkdown. Auto-detects: table for TTY, JSON for pipes
- **Field selection:** `--fields id,name` filters JSON output to specified fields (works on objects and lists)
- **Field presets:** Each list/get command registers a default field preset via `cmdutil.SetFieldsPreset`; `--fields default` resolves it
- **Agent mode:** `--agent` / `RC_AGENT=1` forces compact JSON + default preset + suppresses next-step hints for programmatic use
- **Next-step hints:** Successful mutations can print a stderr-only `next:` hint; suppress with `--no-hints`, `RC_NO_HINTS`, or `--quiet`
- **API client:** All requests go through `internal/api/client.go`. Retries with Retry-After header + backoff_ms + exponential fallback. Verbose HTTP logging, dry-run mode, response caching
- **Error handling:** API errors parsed into `api.Error` struct with type, message, doc_url, StatusCode. Exit codes map HTTP status granularly
- **Confirmation prompts:** All 13 destructive ops (delete/refund/cancel/revoke) require confirmation. `--yes` skips, non-TTY without `--yes` errors safely
- **Interactive create flows:** Create/setup commands prompt for missing required values when stdout is a TTY and return `missing required value` errors in non-interactive scripts
- **Pagination:** Cursor-based with `next_page`. --all and --limit on all 14 list commands
- **Log levels:** `--log-level error|warn|info|debug`, `--verbose` shorthand for debug, `--quiet` suppresses non-essential output
- **Dynamic completions:** Tab-complete resource IDs from cached API responses (8 resource types)
- **No import cycles:** Command packages import `internal/cmdutil` and `internal/completions`, not `cmd`
- **Archive pattern:** Products, entitlements, offerings, and virtual currencies all support archive/unarchive
- **Stability labels:** Commands annotated as [experimental] or [beta] via Cobra annotations

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
make build            # Build with version injection
make test             # Run tests
make test-integration # Run integration tests (requires RC_INTEGRATION_KEY)
make lint             # Run golangci-lint (gofumpt + linters)
make security         # Run gosec security scanner
make check            # Run all checks (fmt, vet, lint, test)
make tools            # Install dev dependencies (gofumpt, gosec, golangci-lint, govulncheck)
make help             # Show all targets
```

## Full API Coverage (104 subcommands, 95 API endpoints)
- [x] Auth: login, status, logout
- [x] Projects: list, create, set-default
- [x] Apps: list, get, create, update, delete, public-keys, storekit-config
- [x] Products: list, get, create, update, delete, archive, unarchive, push-to-store
- [x] Entitlements: list, get, create, update, delete, archive, unarchive, products, attach, detach
- [x] Workflow commands: setup product, offerings publish, paywalls validate, migrate project dry-run
- [x] Offerings: list, get, create, update, delete, publish, archive, unarchive
- [x] Packages: list, get, create, update, delete, products, attach, detach
- [x] Customers: list, lookup, diagnose, create, delete, entitlements, subscriptions, purchases, aliases, attributes, set-attributes, grant, revoke, assign-offering, transfer, restore-purchase, invoices, invoice-file
- [x] Subscriptions: list, get, transactions, entitlements, cancel, refund, refund-transaction, management-url
- [x] Purchases: list, get, entitlements, refund
- [x] Webhooks: list, get, create, update, delete
- [x] Charts: overview, show, options
- [x] Paywalls: list, get, create, delete, validate
- [x] Audit Logs: list
- [x] Collaborators: list
- [x] Virtual Currencies: list, get, create, update, delete, archive, unarchive, balance, credit, set-balance
- [x] JSON + table output (TTY-aware: table for terminal, JSON for pipes)
- [x] Version command with ldflags injection (rc version)
- [x] "Did you mean?" fuzzy command suggestions
- [x] Granular exit codes: 4xx -> 10+(status-400), 5xx -> 60+(status-500)
- [x] JSON + table + markdown output (TTY-aware: table for terminal, JSON for pipes)
- [x] --fields flag for JSON field selection
- [x] --pretty flag, --no-color, NO_COLOR env var
- [x] --verbose / --log-level (error/warn/info/debug) / --quiet
- [x] --dry-run for previewing mutations
- [x] --yes / -y for skipping confirmation prompts on destructive ops
- [x] Confirmation prompts on all 13 destructive operations
- [x] Shell completions: bash/zsh/fish/powershell with dynamic resource ID completion
- [x] Update checking (non-blocking, 24h cache, GitHub releases)
- [x] Local response caching (~/.rc/cache/, 5min TTL)
- [x] MCP server (rc mcp serve) - 15 tools via official Go MCP SDK [experimental]
- [x] Pagination: --all and --limit on all 14 list commands
- [x] Multi-profile auth: --profile flag, RC_PROFILE env, TOML config with migration
- [x] Watch mode: --watch on customers lookup/entitlements, subscriptions get, charts overview
- [x] Bulk import/export: products and entitlements (CSV + JSON)
- [x] Interactive mode: create/setup commands prompt for missing required values when TTY
- [x] Project config transfer: rc export / rc import / rc migrate project --dry-run [beta]
- [x] Strict token validation (sk_/atk_ prefix + length check)
- [x] Config validation on save (profile name format)
- [x] Input validation helpers (internal/validate)
- [x] Stability labels on commands ([experimental], [beta])
- [x] 616 default tests across 41 test files (620 with integration tag)
- [x] Integration test framework (//go:build integration, gated on RC_INTEGRATION_KEY)
- [x] Makefile with tools, security, test-integration targets
- [x] gofumpt formatting, golangci-lint v2, pre-commit hook
- [x] CI: build (Go 1.25 + stable), lint, gosec, govulncheck, CodeQL
- [x] GoReleaser with Homebrew tap + macOS Developer ID signing/notarization
- [x] Install script (curl | sh)
- [x] README with badges, TOC, workflows, troubleshooting
- [x] Community files: SECURITY.md, CODE_OF_CONDUCT.md, SUPPORT.md

## Release
- **v0.1.0** released 2026-04-13 - initial release, 100% API coverage
- **v0.2.0** - MCP server, pagination, multi-profile, watch, bulk ops, interactive mode
- **v0.3.0** - workflow commands, launch diagnostics, customer diagnosis, app credentials, watch expansion, correctness hardening
- Homebrew: `brew install andresdefi/tap/rc`
- Install script: `curl -fsSL https://raw.githubusercontent.com/andresdefi/Revenue-Cat-CLI/main/install.sh | sh`
- Go: `go install github.com/andresdefi/rc@latest`

## New Dependencies (v0.2.0)
- `github.com/modelcontextprotocol/go-sdk` - MCP server
- `github.com/charmbracelet/huh` - interactive prompts
- `github.com/jszwec/csvutil` - CSV marshal/unmarshal
- `github.com/BurntSushi/toml` - TOML config with profiles

## Future Improvements
- [ ] Command grouping in help output (by category)
- [ ] Documentation website (generated from cobra/doc)
- [ ] Set up HOMEBREW_TAP_TOKEN secret for auto formula updates on release

## Module Path
`github.com/andresdefi/rc`

When the GitHub repo name is `Revenue-Cat-CLI`, the module path stays `github.com/andresdefi/rc` for clean imports. The `go install` command uses the module path.
