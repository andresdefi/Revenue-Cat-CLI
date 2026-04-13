# rc - RevenueCat CLI

An unofficial command-line interface for the [RevenueCat REST API v2](https://www.revenuecat.com/docs/api-v2) with **full API coverage**. Manage your projects, products, entitlements, offerings, customers, subscriptions, and more from the terminal.

## Install

### Homebrew (macOS/Linux)

```bash
brew install andresdefi/tap/rc
```

### Go

```bash
go install github.com/andresdefi/rc@latest
```

### Binary releases

Download pre-built binaries from [GitHub Releases](https://github.com/andresdefi/Revenue-Cat-CLI/releases).

## Quick start

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

# Manage entitlements
rc entitlements create --lookup-key premium --display-name "Premium"
rc entitlements attach --entitlement-id entla1b2c3d4e5 --product-id prod1a2b3c4d5e

# Look up a customer
rc customers lookup "user-123"
rc customers entitlements "user-123"
rc customers subscriptions "user-123"

# View metrics
rc charts overview
rc charts show revenue

# Grant an entitlement to a customer
rc customers grant --customer-id user-123 --entitlement-id entla1b2c3 --expires-at 1735689600000
```

## Commands

### Authentication

| Command | Description |
|---------|-------------|
| `rc auth login` | Authenticate with API key |
| `rc auth status` | Show authentication status |
| `rc auth logout` | Remove stored credentials |

### Projects & Apps

| Command | Description |
|---------|-------------|
| `rc projects list/create/set-default` | Manage projects |
| `rc apps list/get/create/update/delete` | Manage platform apps |
| `rc collaborators list` | View project collaborators |

### Product Configuration

| Command | Description |
|---------|-------------|
| `rc products list/get/create/update/delete/archive/unarchive` | Manage products |
| `rc entitlements list/get/create/update/delete/archive/unarchive` | Manage entitlements |
| `rc entitlements products/attach/detach` | Manage entitlement-product associations |
| `rc offerings list/get/create/update/delete/archive/unarchive` | Manage offerings |
| `rc packages list/get/create/update/delete` | Manage packages in offerings |
| `rc packages products/attach/detach` | Manage package-product associations |

### Customer Data

| Command | Description |
|---------|-------------|
| `rc customers list/lookup/create/delete` | Manage customers |
| `rc customers entitlements/subscriptions/purchases/aliases/attributes` | View customer data |
| `rc customers set-attributes` | Set customer attributes |
| `rc customers grant/revoke` | Grant or revoke entitlements |
| `rc customers assign-offering/transfer` | Assign offerings, transfer subscriptions |

### Subscriptions & Purchases

| Command | Description |
|---------|-------------|
| `rc subscriptions list/get` | View subscriptions |
| `rc subscriptions transactions/entitlements` | View subscription details |
| `rc subscriptions cancel/refund/refund-transaction` | Cancel or refund |
| `rc subscriptions management-url` | Get authenticated management URL |
| `rc purchases list/get/entitlements/refund` | View and refund purchases |

### Integrations & Analytics

| Command | Description |
|---------|-------------|
| `rc webhooks list/get/create/update/delete` | Manage webhooks |
| `rc charts overview` | View metrics overview |
| `rc charts show <chart-name>` | View specific chart data |
| `rc paywalls list/get/delete` | Manage paywalls |
| `rc audit-logs list` | View audit logs |

### Virtual Currencies

| Command | Description |
|---------|-------------|
| `rc currencies list/get/create/update/delete/archive/unarchive` | Manage currencies |
| `rc currencies balance` | View customer balances |
| `rc currencies credit/set-balance` | Credit or set balances |

## Global flags

| Flag | Short | Description |
|------|-------|-------------|
| `--project` | `-p` | Project ID (overrides default) |
| `--output` | `-o` | Output format: `table` (default), `json` |

## Authentication

`rc` uses RevenueCat API v2 secret keys (prefixed `sk_`). Create one in the RevenueCat dashboard under **Project Settings > API Keys > + New Secret API Key**. Make sure to enable v2 API permissions.

Your key is stored in the system keychain (macOS Keychain, Windows Credential Manager, or Linux Secret Service). If no keychain is available, it falls back to `~/.rc/config.json`.

## Output formats

All commands support both table and JSON output:

```bash
# Pretty table (default)
rc products list

# JSON (for scripting/piping)
rc products list -o json

# Pipe to jq
rc products list -o json | jq '.items[].store_identifier'
```

## Shell completion

```bash
# Bash
rc completion bash > /etc/bash_completion.d/rc

# Zsh
rc completion zsh > "${fpath[1]}/_rc"

# Fish
rc completion fish > ~/.config/fish/completions/rc.fish
```

## API coverage

This CLI provides **full coverage** of the RevenueCat REST API v2 with 91 subcommands across 16 command groups: projects, apps, products, entitlements, offerings, packages, customers, subscriptions, purchases, webhooks, charts, paywalls, audit logs, collaborators, and virtual currencies.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE)

## Disclaimer

This is an unofficial, community-maintained tool. It is not affiliated with or endorsed by RevenueCat, Inc.
