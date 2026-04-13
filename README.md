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

# Set a default project (so you don't need --project every time)
rc projects set-default proj1ab2c3d4

# List products
rc products list

# Create a product
rc products create --store-id "com.app.premium_monthly" --app-id app1a2b3c4 --type subscription

# Look up a customer
rc customers lookup "user-123"

# View metrics
rc charts overview
```

## Commands

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
| `rc entitlements list/get/create/update/delete` | Manage entitlements |
| `rc entitlements archive/unarchive` | Archive or restore entitlements |
| `rc entitlements products/attach/detach` | Manage product associations |
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

Your key is stored in the system keychain (macOS Keychain, Windows Credential Manager, or Linux Secret Service). If no keychain is available, it falls back to `~/.rc/config.json`.

```bash
rc auth login     # Enter your API key
rc auth status    # Check current auth
rc auth logout    # Remove stored key
```

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

**100% coverage** of the RevenueCat REST API v2 - all 95 endpoints across 16 resource groups. Verified against the [official OpenAPI spec](https://www.revenuecat.com/docs/api-v2).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines. Run `make check` before submitting PRs.

## License

[MIT](LICENSE)

## Disclaimer

This is an unofficial, community-maintained tool. It is not affiliated with or endorsed by RevenueCat, Inc. RevenueCat is a trademark of RevenueCat, Inc.
