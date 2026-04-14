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

# Authenticate with a named profile for staging
rc auth login --profile staging

# List your projects
rc projects list

# Set a default project (so you don't need --project every time)
rc projects set-default proj1ab2c3d4

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

All commands support `--profile <name>` to select the config profile and `--output json|table` to control output format.

### Authentication

| Command | Description |
|---------|-------------|
| `rc auth login` | Authenticate with API key |
| `rc auth status` | Show authentication status |
| `rc auth logout` | Remove stored credentials |
| `rc version` | Print version info |

### Projects & Apps

| Command | Description |
|---------|-------------|
| `rc projects list/create/set-default` | Manage projects |
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
| `rc export --file config.json` | Export full project config (products, entitlements, offerings) |
| `rc import --file config.json` | Import project config from a JSON file |

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
# List all product IDs
rc products list -o json | jq -r '.items[].id'

# Find active subscribers
rc customers list -o json | jq '.items[] | select(.active_entitlements.items | length > 0)'

# Get revenue overview
rc charts overview -o json | jq '.metrics[] | {name, value}'
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
# Export full project config (products, entitlements, offerings)
rc export --file project-config.json --project proj_source

# Import into a different project
rc import --file project-config.json --project proj_target
```

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

# Force table in pipe
rc products list -o table | head
```

## Authentication

`rc` uses RevenueCat API v2 secret keys (prefixed `sk_`). Create one in the RevenueCat dashboard under **Project Settings > API Keys > + New Secret API Key**. Make sure to enable v2 API permissions.

Your key is stored in the system keychain (macOS Keychain, Windows Credential Manager, or Linux Secret Service). If no keychain is available, it falls back to the config file.

```bash
rc auth login     # Enter your API key
rc auth status    # Check current auth
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

```bash
# Bash
rc completion bash > /etc/bash_completion.d/rc

# Zsh
rc completion zsh > "${fpath[1]}/_rc"

# Fish
rc completion fish > ~/.config/fish/completions/rc.fish
```

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

All RevenueCat API errors include a `doc_url` with details. Use `--output json` to see the full error response:

```bash
rc products create --store-id test --app-id bad --type sub -o json
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
