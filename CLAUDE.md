# rc - RevenueCat CLI

## Project Overview
Unofficial open-source CLI for the RevenueCat REST API v2. Written in Go with Cobra.
Repo: `andresdefi/Revenue-Cat-CLI` on GitHub. Binary name: `rc`.
100% API v2 coverage - 99 subcommands covering all 95 API endpoints across 16 command groups.

## Tech Stack
- **Language:** Go 1.22+
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
go build -o rc .
./rc --help
./rc auth login
./rc projects list
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
- [x] JSON + table output
- [x] README, LICENSE (MIT), CONTRIBUTING.md
- [x] GitHub Actions CI (build, test, lint)
- [x] GoReleaser config with Homebrew tap

## Future Improvements
- [ ] Pagination support (auto-fetch all pages with `--all` flag)
- [ ] Version command (`rc version` with ldflags injection)
- [ ] Interactive mode for create commands (prompt for required fields)
- [ ] `--watch` flag for polling commands
- [ ] Unit tests for API client and output formatting
- [ ] Integration test suite with mock server

## Module Path
`github.com/andresdefi/rc`

When the GitHub repo name is `Revenue-Cat-CLI`, the module path stays `github.com/andresdefi/rc` for clean imports. The `go install` command uses the module path.
