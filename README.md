# rc - RevenueCat CLI

[![Release](https://img.shields.io/github/v/release/andresdefi/Revenue-Cat-CLI?style=for-the-badge)](https://github.com/andresdefi/Revenue-Cat-CLI/releases)
[![Go](https://img.shields.io/github/go-mod/go-version/andresdefi/Revenue-Cat-CLI?style=for-the-badge)](https://go.dev/)
[![License](https://img.shields.io/github/license/andresdefi/Revenue-Cat-CLI?style=for-the-badge)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/andresdefi/Revenue-Cat-CLI/ci.yaml?branch=main&style=for-the-badge&label=CI)](https://github.com/andresdefi/Revenue-Cat-CLI/actions/workflows/ci.yaml)

An unofficial command-line interface for the [RevenueCat REST API v2](https://www.revenuecat.com/docs/api-v2) with **100% API coverage** (95 endpoints). Manage your projects, products, entitlements, offerings, customers, subscriptions, and more from the terminal.

## Table of Contents

- [Install](#install)
- [Quick Start](#quick-start)
- [Commands](#commands)
- [Common Workflows](#common-workflows)
- [Documentation](#documentation)
- [Output Formats](#output-formats)
- [Authentication](#authentication)
- [Profiles](#profiles)
- [MCP Server](#mcp-server)
- [Shell Completion](#shell-completion)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)

## Install

### Homebrew (macOS/Linux)

```bash
brew install andresdefi/tap/rc
```

### Install script

```bash
curl -fsSL https://raw.githubusercontent.com/andresdefi/Revenue-Cat-CLI/main/install.sh | sh
```

### Go

```bash
go install github.com/andresdefi/rc@latest
```

### Binary releases

Download pre-built binaries from [GitHub Releases](https://github.com/andresdefi/Revenue-Cat-CLI/releases).

## Quick Start

```bash
# Authenticate with your RevenueCat API v2 secret key
rc auth login

# List your projects
rc projects list

# Initialize config and set a default project
rc init --project proj1ab2c3d4

# Validate local config, auth, default project, and API access
rc doctor

# Check RevenueCat project setup before launch
rc project doctor

# Run a stricter launch readiness preflight
rc launch-check

# Confirm the active profile context
rc whoami

# List products (all pages)
rc products list --all

# Create a product
rc products create --store-id "com.app.premium_monthly" --app-id app1a2b3c4 --type subscription

# Look up a customer with live refresh
rc customers lookup "user-123" --watch

# View metrics
rc charts overview

# Start MCP server for Claude Code integration
rc mcp serve
```

## Commands

All commands support `--profile <name>` to select the config profile and `--output json|table|markdown` to control output format. The generated command reference is in [docs/COMMANDS.md](docs/COMMANDS.md).

### Foundation

| Command | Description |
|---------|-------------|
| `rc init` | Initialize config for a profile and default project |
| `rc doctor` | Check config, auth, project, and API connectivity |
| `rc whoami` | Show active profile, auth source, and default project |
| `rc config profiles` | List configured profiles |
| `rc launch-check` | Check whether project setup has the required launch paths |

### Authentication

| Command | Description |
|---------|-------------|
| `rc auth login` | Authenticate with API key |
| `rc auth status` | Show authentication status |
| `rc auth doctor` | Check authentication health and API connectivity |
| `rc auth validate` | Alias-style validation for auth health |
| `rc auth logout` | Remove stored credentials |
| `rc version` | Print version info |

### Projects & Apps

| Command | Description |
|---------|-------------|
| `rc projects list/create/doctor/set-default` | Manage projects and project health |
| `rc apps list/get/create/update/delete` | Manage platform apps |
| `rc apps public-keys <app-id>` | List public API keys |
| `rc apps storekit-config <app-id>` | Get StoreKit configuration |
| `rc collaborators list` | View project collaborators |

### Product Configuration

| Command | Description |
|---------|-------------|
| `rc products list/get/create/update/delete` | Manage products |
| `rc products archive/unarchive` | Archive or restore products |
| `rc products push-to-store <id>` | Push product to connected store |
| `rc products export --file products.csv` | Export products to CSV or JSON |
| `rc products import --file products.csv` | Import products from CSV or JSON |
| `rc entitlements list/get/create/update/delete` | Manage entitlements |
| `rc entitlements archive/unarchive` | Archive or restore entitlements |
| `rc entitlements products/attach/detach` | Manage product associations |
| `rc entitlements export --file ent.csv` | Export entitlements to CSV or JSON |
| `rc entitlements import --file ent.csv` | Import entitlements from CSV or JSON |
| `rc offerings list/get/create/update/delete` | Manage offerings |
| `rc offerings archive/unarchive` | Archive or restore offerings |
| `rc packages list/get/create/update/delete` | Manage packages |
| `rc packages products/attach/detach` | Manage package products |

### Customer Data

| Command | Description |
|---------|-------------|
| `rc customers list/lookup/create/delete` | Manage customers |
| `rc customers entitlements <id>` | Active entitlements |
| `rc customers subscriptions <id>` | Customer subscriptions |
| `rc customers purchases <id>` | Customer purchases |
| `rc customers aliases/attributes <id>` | Aliases and attributes |
| `rc customers set-attributes` | Set customer attributes |
| `rc customers grant/revoke` | Grant or revoke entitlements |
| `rc customers assign-offering/transfer` | Assign offerings, transfer data |
| `rc customers restore-purchase` | Restore Google Play purchase |
| `rc customers invoices/invoice-file` | View and download invoices |

### Subscriptions & Purchases

| Command | Description |
|---------|-------------|
| `rc subscriptions list/get` | View subscriptions |
| `rc subscriptions transactions/entitlements` | Subscription details |
| `rc subscriptions cancel/refund` | Cancel or refund |
| `rc subscriptions refund-transaction` | Refund specific transaction |
| `rc subscriptions management-url` | Get management portal URL |
| `rc purchases list/get/entitlements/refund` | View and refund purchases |

### Integrations & Analytics

| Command | Description |
|---------|-------------|
| `rc webhooks list/get/create/update/delete` | Manage webhooks |
| `rc charts overview` | Metrics overview |
| `rc charts show <name>` | Chart data (revenue, mrr, etc.) |
| `rc charts options <name>` | Chart filter options |
| `rc paywalls list/get/create/delete` | Manage paywalls |
| `rc audit-logs list` | View audit logs |

### Virtual Currencies

| Command | Description |
|---------|-------------|
| `rc currencies list/get/create/update/delete` | Manage currencies |
| `rc currencies archive/unarchive` | Archive or restore |
| `rc currencies balance` | Customer balances |
| `rc currencies credit/set-balance` | Credit or set balances |

### Data Transfer

| Command | Description |
|---------|-------------|
| `rc export --file config.json` | Export full project config (apps, products, entitlements, offerings, packages, attachments) |
| `rc import --file config.json` | Import project config into another project |

### MCP

| Command | Description |
|---------|-------------|
| `rc mcp serve` | Start MCP server over stdio |

### Flags on List Commands

All list commands support pagination:

| Flag | Description |
|------|-------------|
| `--all` | Fetch all pages (follows cursor pagination) |
| `--limit N` | Fetch up to N items |

### Flags on Lookup/Get Commands

Select commands support live refresh:

| Flag | Description |
|------|-------------|
| `--watch` | Auto-refresh every 5 seconds (Ctrl+C to stop) |

Supported on: `rc customers lookup`, `rc customers entitlements`, `rc subscriptions get`, `rc charts overview`.

## Common Workflows

More recipes live in [docs/WORKFLOWS.md](docs/WORKFLOWS.md).

### Set up a new product with entitlements

```bash
# Create the product
rc products create --store-id "com.app.pro_monthly" --app-id app1a2b3c4 --type subscription

# Create an entitlement
rc entitlements create --lookup-key pro --display-name "Pro Access"

# Attach the product to the entitlement
rc entitlements attach --entitlement-id entla1b2c3d4e5 --product-id prod1a2b3c4d5e
```

### Set up an offering with packages

```bash
# Create an offering
rc offerings create --lookup-key default --display-name "Default Offering"

# Create packages in the offering
rc packages create --offering-id ofrnge1a2b3c --lookup-key monthly --display-name "Monthly"
rc packages create --offering-id ofrnge1a2b3c --lookup-key annual --display-name "Annual" --position 2

# Attach products to packages
rc packages attach --package-id pkge1a2b3c --product-id prod1a2b3c
```

### Debug a customer's subscription

```bash
# Look up the customer
rc customers lookup "user-123"

# Check their active entitlements
rc customers entitlements "user-123"

# List their subscriptions
rc customers subscriptions "user-123"

# Get subscription details
rc subscriptions get sub1ab2c3d4e5
```

### Grant promotional access

```bash
# Grant entitlement (expires in 30 days)
rc customers grant \
  --customer-id "user-123" \
  --entitlement-id entla1b2c3 \
  --expires-at 1738281600000

# Later, revoke if needed
rc customers revoke --customer-id "user-123" --entitlement-id entla1b2c3
```

### Script with JSON output

```bash
# List all product IDs (using --fields for clean output)
rc products list -o json --fields id,store_identifier

# Or use jq for more complex filtering
rc products list -o json | jq -r '.items[].id'

# Find active subscribers
rc customers list -o json | jq '.items[] | select(.active_entitlements.items | length > 0)'

# Get revenue overview
rc charts overview -o json | jq '.metrics[] | {name, value}'

# Quiet mode for scripts (only data, no status messages)
rc products delete prod1a2b3c --yes --quiet
```

### Multi-profile workflow

```bash
# Set up profiles for different environments
rc auth login --profile prod
rc auth login --profile staging

# Use staging profile for testing
rc products list --profile staging

# Switch back to production
rc products list --profile prod

# Set the default profile in config
# (edit ~/.rc/config.toml and set current_profile)
```

### Bulk import/export workflow

```bash
# Export products and entitlements to CSV for review
rc products export --file products.csv
rc entitlements export --file entitlements.csv

# Edit the CSV files as needed, then import into another project
rc products import --file products.csv --project proj_target123
rc entitlements import --file entitlements.csv --project proj_target123
```

### Project migration (export + import)

```bash
# Export full project config, including relationships and attachments
rc export --file project-config.json --project proj_source

# Import into a different project
rc import --file project-config.json --project proj_target

# Map source apps to target apps when automatic matching is not enough
rc import --file project-config.json --project proj_target --app-map app_source=app_target
```

`rc export` and `rc import` are beta. Run with `--dry-run` before applying a migration to another project.

## Documentation

| Document | Purpose |
|----------|---------|
| [docs/COMMANDS.md](docs/COMMANDS.md) | Generated command reference from Cobra help |
| [docs/WORKFLOWS.md](docs/WORKFLOWS.md) | Copyable RevenueCat workflow recipes |
| [docs/API_NOTES.md](docs/API_NOTES.md) | API semantics, pagination, transfer, and error notes |
| [docs/CI_CD.md](docs/CI_CD.md) | CI/CD setup and non-interactive usage |
| [docs/TESTING.md](docs/TESTING.md) | Fixture, golden request, pagination, and integration test guidance |

### MCP server setup

```bash
# Start the MCP server (communicates over stdio)
rc mcp serve
```

To use with Claude Code, add to your `.claude/settings.json`:

```json
{
  "mcpServers": {
    "revenuecat": {
      "command": "rc",
      "args": ["mcp", "serve"]
    }
  }
}
```

The MCP server exposes 16 tools for RevenueCat operations: list/get/create for projects, products, entitlements, offerings, customers, subscriptions, plus customer entitlement grants and metrics overview. It uses the same authentication as the CLI (RC_API_KEY env var or stored keychain token).

## Output Formats

Output adapts to context automatically:

| Context | Default | Override |
|---------|---------|---------|
| Terminal (TTY) | Table | `rc products list -o json` |
| Pipe / redirect | JSON | `rc products list -o table` |

```bash
# Pretty table in terminal
rc products list

# JSON for scripting (auto when piped)
rc products list | jq '.items[].id'

# Force JSON in terminal
rc products list -o json

# Markdown table (great for docs and GitHub issues)
rc products list -o markdown

# Select specific fields in JSON output
rc products list -o json --fields id,store_identifier,state

# Pretty-print JSON even when piping
rc products list -o json --pretty | less
```

## Global Flags

These flags work on all commands:

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output format: `table`, `json`, `markdown` (default: table for TTY, json for pipes) |
| `--project` | `-p` | Project ID (overrides default) |
| `--profile` | | Config profile to use |
| `--fields` | | Comma-separated field selection for JSON output |
| `--pretty` | | Pretty-print JSON (default for TTY, compact for pipes) |
| `--verbose` | `-v` | Shorthand for `--log-level debug` |
| `--log-level` | | Log verbosity: `error`, `warn`, `info`, `debug` (default: `warn`) |
| `--quiet` | `-q` | Suppress success messages, warnings, and progress output |
| `--dry-run` | | Preview mutations without executing (shows method, path, body) |
| `--yes` | `-y` | Skip confirmation prompts on destructive operations |
| `--no-color` | | Disable color output (also respects `NO_COLOR` env var) |

### Destructive operations

Delete, refund, cancel, and revoke commands prompt for confirmation before executing. Use `--yes` to skip in scripts:

```bash
# Interactive: prompts "Delete product prod1a2b3c4d5?"
rc products delete prod1a2b3c4d5

# Non-interactive: skips the prompt
rc products delete prod1a2b3c4d5 --yes

# Preview what would happen without doing it
rc products delete prod1a2b3c4d5 --dry-run
```

### Debugging API issues

```bash
# See HTTP request/response details
rc products list --verbose

# Or use tiered log levels
rc products list --log-level debug
```

## Authentication

`rc` uses RevenueCat API v2 secret keys (prefixed `sk_`). Create one in the RevenueCat dashboard under **Project Settings > API Keys > + New Secret API Key**. Make sure to enable v2 API permissions.

Your key is stored in the system keychain (macOS Keychain, Windows Credential Manager, or Linux Secret Service). If no keychain is available, it falls back to the config file.

```bash
rc auth login     # Enter your API key
rc auth status    # Check current auth
rc auth validate  # Validate API connectivity
rc auth logout    # Remove stored key
```

## Profiles

rc supports multiple profiles for different environments (production, staging, development, etc.):

```bash
# Log in with a named profile
rc auth login --profile prod
rc auth login --profile staging

# Use a specific profile for any command
rc products list --profile staging

# Check auth status for a profile
rc auth status --profile staging
```

### Profile resolution order

1. `--profile` flag (highest priority)
2. `RC_PROFILE` environment variable
3. `current_profile` setting in `~/.rc/config.toml`
4. Falls back to `default` profile

### Config file

Profiles are stored in `~/.rc/config.toml`:

```toml
current_profile = "default"

[profiles.default]
api_key = "sk_..."
project_id = "proj_abc123"

[profiles.staging]
api_key = "sk_..."
project_id = "proj_staging456"
```

If you previously used rc v0.1.0, the old `~/.rc/config.json` is automatically migrated to the new TOML format on first run.

## MCP Server

rc includes a built-in [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server that exposes RevenueCat operations as tools for AI assistants.

```bash
rc mcp serve
```

The server communicates over stdin/stdout using the MCP protocol.

### Claude Code integration

Add to your `.claude/settings.json`:

```json
{
  "mcpServers": {
    "revenuecat": {
      "command": "rc",
      "args": ["mcp", "serve"]
    }
  }
}
```

### Available tools

The MCP server exposes 16 tools:

- **Projects**: `list_projects`
- **Products**: `list_products`, `get_product`, `create_product`
- **Entitlements**: `list_entitlements`, `get_entitlement`, `create_entitlement`
- **Offerings**: `list_offerings`, `get_offering`
- **Customers**: `lookup_customer`, `list_customer_entitlements`, `grant_entitlement`
- **Subscriptions**: `list_subscriptions`, `get_subscription`
- **Metrics**: `metrics_overview`

### Authentication

The MCP server resolves API keys in this order:

1. `RC_API_KEY` environment variable
2. Stored keychain / config token (same as normal CLI auth)

## Shell Completion

Generate completions for your shell:

```bash
# Bash
source <(rc completion bash)
# Or persist: rc completion bash > /etc/bash_completion.d/rc

# Zsh
rc completion zsh > "${fpath[1]}/_rc"

# Fish
rc completion fish > ~/.config/fish/completions/rc.fish

# PowerShell
rc completion powershell | Out-String | Invoke-Expression
```

Completions include **dynamic resource ID completion** - tab-complete product IDs, entitlement IDs, offering IDs, and more directly from the API (cached locally for speed). Works for products, entitlements, offerings, packages, apps, webhooks, subscriptions, and projects.

## Exit Codes

Exit codes map HTTP status codes for scripting:

| Exit code | Meaning |
|-----------|---------|
| `0` | Success |
| `1` | General error |
| `2` | Usage error (bad flags/args) |
| `3` | Config error |
| `10-49` | HTTP 4xx: `10 + (status - 400)` (e.g. 401 -> 11, 404 -> 14, 429 -> 39) |
| `60-99` | HTTP 5xx: `60 + (status - 500)` (e.g. 500 -> 60, 503 -> 63) |

```bash
rc products get bad-id; echo $?  # 14 (404 Not Found)
```

## Update Notifications

rc checks for new versions in the background (once every 24 hours). If a newer release is available, a notice appears after the command completes. This never blocks or slows down your commands.

## Troubleshooting

### "not logged in"

```bash
rc auth status    # Check if authenticated
rc auth login     # Re-authenticate
```

### "no project specified"

```bash
rc projects list                        # Find your project ID
rc projects set-default proj1ab2c3d4    # Set default
```

### API errors

All RevenueCat API errors include a `doc_url` with details. Use `--verbose` to see full HTTP request/response details:

```bash
rc products create --store-id test --app-id bad --type sub --verbose
```

### Rate limits

If you hit `rate_limit_error`, wait and retry. Limits:
- Project configuration: 60 req/min
- Customer information: 480 req/min
- Charts: 5 req/min

### Homebrew issues

```bash
brew update
brew reinstall andresdefi/tap/rc
```

## API Coverage

**100% coverage** of the RevenueCat REST API v2 - all 95 endpoints across 16 resource groups, plus 110+ subcommands including bulk operations, data transfer, and MCP tools. Verified against the [official OpenAPI spec](https://www.revenuecat.com/docs/api-v2).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines. Run `make check` before submitting PRs.

## License

[MIT](LICENSE)

## Disclaimer

This is an unofficial, community-maintained tool. It is not affiliated with or endorsed by RevenueCat, Inc. RevenueCat is a trademark of RevenueCat, Inc.
