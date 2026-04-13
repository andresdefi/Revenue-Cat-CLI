# rc - RevenueCat CLI

An unofficial command-line interface for the [RevenueCat REST API v2](https://www.revenuecat.com/docs/api-v2). Manage your RevenueCat projects, products, entitlements, offerings, and customers from the terminal.

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

Download pre-built binaries from [GitHub Releases](https://github.com/andresdefi/rc/releases).

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

# List entitlements
rc entitlements list

# Create an entitlement and attach a product
rc entitlements create --lookup-key premium --display-name "Premium"
rc entitlements attach --entitlement-id entla1b2c3d4e5 --product-id prod1a2b3c4d5e

# List offerings
rc offerings list

# Look up a customer
rc customers lookup "user-123"
rc customers entitlements "user-123"
```

## Commands

| Command | Description |
|---------|-------------|
| `rc auth login` | Authenticate with API key |
| `rc auth status` | Show authentication status |
| `rc auth logout` | Remove stored credentials |
| `rc projects list` | List all projects |
| `rc projects set-default` | Set default project |
| `rc products list` | List products |
| `rc products create` | Create a product |
| `rc entitlements list` | List entitlements |
| `rc entitlements create` | Create an entitlement |
| `rc entitlements attach` | Attach products to entitlement |
| `rc entitlements detach` | Detach products from entitlement |
| `rc offerings list` | List offerings |
| `rc offerings create` | Create an offering |
| `rc customers lookup` | Look up customer by ID |
| `rc customers entitlements` | List customer's active entitlements |

## Global flags

| Flag | Short | Description |
|------|-------|-------------|
| `--project` | `-p` | Project ID (overrides default) |
| `--output` | `-o` | Output format: `table` (default), `json` |

## Authentication

`rc` uses RevenueCat API v2 secret keys (prefixed `sk_`). Create one in the RevenueCat dashboard under **Project Settings > API Keys > + New Secret API Key**. Make sure to enable v2 API permissions.

Your key is stored in the system keychain (macOS Keychain, Windows Credential Manager, or Linux Secret Service). If no keychain is available, it falls back to `~/.rc/config.json`.

## Output formats

All list commands support both table and JSON output:

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

This CLI covers the core RevenueCat API v2 endpoints. The full API also supports apps, packages, subscriptions, purchases, webhooks, charts, and more. Contributions welcome!

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE)

## Disclaimer

This is an unofficial, community-maintained tool. It is not affiliated with or endorsed by RevenueCat, Inc.
