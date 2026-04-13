# rc - RevenueCat CLI

## Project Overview
Unofficial open-source CLI for the RevenueCat REST API v2. Written in Go with Cobra.
Repo: `andresdefi/Revenue-Cat-CLI` on GitHub. Binary name: `rc`.

## Tech Stack
- **Language:** Go 1.22+
- **CLI framework:** Cobra (github.com/spf13/cobra)
- **Table output:** go-pretty/v6 (github.com/jedib0t/go-pretty/v6)
- **Keychain:** go-keyring (github.com/zalando/go-keyring)
- **Release:** GoReleaser, Homebrew tap
- **CI:** GitHub Actions (build, test, lint, release)

## Architecture
```
main.go                    Entry point
cmd/
  root.go                  Root command, persistent flags (--project, --output)
  auth/auth.go             auth login/status/logout
  projects/projects.go     projects list/set-default
  products/products.go     products list/create
  entitlements/entitlements.go  entitlements list/create/attach/detach
  offerings/offerings.go   offerings list/create
  customers/customers.go   customers lookup/entitlements
internal/
  api/client.go            HTTP client with retry, auth header, error handling
  api/types.go             All RevenueCat API v2 response types
  auth/auth.go             Token storage (keychain + config fallback)
  config/config.go         Config file (~/.rc/config.json)
  cmdutil/cmdutil.go       Shared helpers (ResolveProject, GetOutputFormat)
  output/output.go         JSON + table output formatting
```

## Key Patterns
- **Auth:** Bearer token via `Authorization: Bearer <sk_...>`. Stored in system keychain, falls back to `~/.rc/config.json`
- **Project resolution:** `--project` flag > config default > error with guidance
- **Output:** `--output table` (default) or `--output json`. Table uses go-pretty, JSON uses encoding/json
- **API client:** All requests go through `internal/api/client.go`. Retries on `retryable: true` errors with backoff
- **Error handling:** API errors parsed into `api.Error` struct with type, message, doc_url
- **Pagination:** Cursor-based with `starting_after` param (not yet implemented in list commands - lists return first page)
- **No import cycles:** Command packages import `internal/cmdutil`, not `cmd`

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

## What's Implemented
- [x] Auth (login, status, logout) with keychain + config fallback
- [x] Projects (list, set-default)
- [x] Products (list, create)
- [x] Entitlements (list, create, attach, detach)
- [x] Offerings (list, create)
- [x] Customers (lookup, entitlements)
- [x] JSON + table output
- [x] README, LICENSE (MIT), CONTRIBUTING.md
- [x] GitHub Actions CI (build, test, lint)
- [x] GoReleaser config with Homebrew tap
- [x] GitHub repo created

## What's Next (Future)
- [ ] Pagination support (auto-fetch all pages with `--all` flag)
- [ ] `rc apps list/create` commands
- [ ] `rc packages list/create/attach` commands
- [ ] `rc subscriptions list` and customer subscriptions
- [ ] `rc webhooks list/create/delete` commands
- [ ] `rc charts` for analytics
- [ ] Version command (`rc version` with ldflags injection)
- [ ] Interactive mode for create commands (prompt for required fields)
- [ ] `--watch` flag for polling commands
- [ ] Unit tests for API client and output formatting
- [ ] Integration test suite with mock server

## Module Path
`github.com/andresdefi/rc`

When the GitHub repo name is `Revenue-Cat-CLI`, the module path stays `github.com/andresdefi/rc` for clean imports. The `go install` command uses the module path.
