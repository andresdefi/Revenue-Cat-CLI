# rc Command Reference

Generated from Cobra command definitions. Do not edit by hand.

## rc

RevenueCat CLI - manage your RevenueCat projects from the terminal

rc is an unofficial CLI for the RevenueCat REST API v2.

Manage products, entitlements, offerings, customers, subscriptions,
and more without leaving your terminal.

Get started:
  rc auth login           Authenticate with your API key
  rc projects list        List your projects
  rc products list        List products in a project
  rc customers lookup     Look up a customer
  rc charts overview      View metrics overview

Full API v2 coverage: projects, apps, products, entitlements, offerings,
packages, customers, subscriptions, purchases, webhooks, charts, paywalls,
audit logs, collaborators, and virtual currencies.

**Flags**

- `--dry-run`: show what would be done without executing mutations Default: `false`.
- `--fields`: comma-separated list of fields to include in JSON output
- `--log-level`: log verbosity: error, warn, info, debug (default: warn)
- `--no-color`: disable color output (also respects NO_COLOR env var) Default: `false`.
- `-o, --output`: output format: table, json, markdown (default: table for TTY, json for pipes)
- `--pretty`: pretty-print JSON output (default for TTY, compact for pipes) Default: `false`.
- `--profile`: config profile to use (overrides RC_PROFILE and current_profile)
- `-p, --project`: project ID (overrides default project)
- `-q, --quiet`: suppress non-essential output (success messages, warnings, progress) Default: `false`.
- `-v, --verbose`: shorthand for --log-level debug Default: `false`.
- `-y, --yes`: skip confirmation prompts for destructive operations Default: `false`.

### rc apps

Manage apps within a project

Manage platform apps within a RevenueCat project.

Each project can have multiple apps for different platforms (iOS, Android, Stripe, etc).
Products are created per-app using the store identifier from the respective platform.

Examples:
  rc apps list
  rc apps get app1a2b3c4
  rc apps create --name "My iOS App" --type app_store --bundle-id com.example.app
  rc apps delete app1a2b3c4

#### rc apps create

Create a new app

Create a new platform app in a project.

Supported types: app_store, play_store, amazon, stripe, rc_billing,
roku, mac_app_store, paddle

**Flags**

- `--bundle-id`: bundle ID / package name (for app_store, play_store, amazon)
- `--name`: app name (required)
- `--type`: platform type: app_store, play_store, amazon, stripe, rc_billing, roku, mac_app_store, paddle (required)

**Examples**

```bash
# Create an iOS app
  rc apps create --name "iOS App" --type app_store --bundle-id com.example.app

  # Create an Android app
  rc apps create --name "Android App" --type play_store --bundle-id com.example.app

  # Create a Stripe app
  rc apps create --name "Web Payments" --type stripe
```

#### rc apps delete

Delete an app

**Examples**

```bash
# Delete an app
  rc apps delete app1a2b3c4d5
```

#### rc apps get

Get an app by ID

**Examples**

```bash
# Get app details
  rc apps get app1a2b3c4d5

  # Get as JSON
  rc apps get app1a2b3c4d5 -o json
```

#### rc apps list

List apps in a project

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List all apps
  rc apps list

  # List with JSON output
  rc apps list -o json

  # Fetch all pages
  rc apps list --all
```

#### rc apps public-keys

List public API keys for an app

**Examples**

```bash
# List public keys for an app
  rc apps public-keys app1a2b3c4d5

  # Get as JSON
  rc apps public-keys app1a2b3c4d5 -o json
```

#### rc apps storekit-config

Get StoreKit configuration for an app

**Examples**

```bash
# Get StoreKit config for an iOS app
  rc apps storekit-config app1a2b3c4d5

  # Get as JSON
  rc apps storekit-config app1a2b3c4d5 -o json
```

#### rc apps update

Update an app

**Flags**

- `--name`: new app name
- `--service-account-file`: path to Google Play service account JSON file (not supported by RevenueCat API v2 app update)
- `--shared-secret`: App Store shared secret
- `--subscription-key-file`: path to App Store in-app purchase key .p8 file
- `--subscription-key-id`: App Store in-app purchase key ID
- `--subscription-key-issuer`: App Store in-app purchase key issuer ID

**Examples**

```bash
# Rename an app
  rc apps update app1a2b3c4d5 --name "My Renamed App"

  # Configure App Store shared secret
  rc apps update app1a2b3c4d5 --shared-secret 1234567890abcdef1234567890abcdef

  # Configure App Store in-app purchase key
  rc apps update app1a2b3c4d5 \
    --subscription-key-file ./SubscriptionKey_ABC123.p8 \
    --subscription-key-id ABC123 \
    --subscription-key-issuer 5a049d62-1b9b-453c-b605-1988189d8129
```

### rc audit-logs

View audit logs

#### rc audit-logs list

List audit log entries

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--end-date`: filter to date (YYYY-MM-DD)
- `--limit`: max items per page Default: `0`.
- `--start-date`: filter from date (YYYY-MM-DD)

**Examples**

```bash
# List recent audit logs
  rc audit-logs list

  # Filter by date range
  rc audit-logs list --start-date 2024-01-01 --end-date 2024-01-31

  # Fetch all pages
  rc audit-logs list --all
```

### rc auth

Manage authentication

Log in, check status, or log out of the RevenueCat API.

#### rc auth doctor

Check authentication health and API connectivity

**Examples**

```bash
# Run auth diagnostics
  rc auth doctor

  # Check a specific profile
  rc auth doctor --profile production

  # Validate auth before listing projects
  rc auth doctor
  rc projects list

  # Diagnose a staging profile
  rc auth status --profile staging
  rc auth doctor --profile staging

  # Use in scripts before a release check
  rc auth doctor --profile production >/dev/null && rc products list --profile production --output json
```

#### rc auth login

Authenticate with a RevenueCat API v2 secret key

Authenticate with your RevenueCat API v2 secret key.

You can create a v2 secret key in the RevenueCat dashboard:
  Project Settings > API Keys > + New Secret API Key

The key will be stored in your system keychain (with config file fallback).
Keys are prefixed with sk_ and must have v2 API permissions.

**Examples**

```bash
# Log in with the default profile
  rc auth login

  # Log in with a specific profile
  rc auth login --profile staging

  # Check the profile after login
  rc auth login --profile production
  rc auth status --profile production

  # Verify API access after saving a key
  rc auth login
  rc auth doctor

  # Switch between profiles for project work
  rc auth login --profile staging
  rc projects list --profile staging
```

#### rc auth logout

Remove stored API key

**Examples**

```bash
# Log out of the default profile
  rc auth logout

  # Log out of a specific profile
  rc auth logout --profile staging
```

#### rc auth status

Show current authentication status

**Examples**

```bash
# Check auth status for the default profile
  rc auth status

  # Check auth status for a specific profile
  rc auth status --profile production

  # Run diagnostics after checking status
  rc auth status
  rc auth doctor

  # Confirm a profile before listing projects
  rc auth status --profile staging
  rc projects list --profile staging

  # Use in scripts before a workflow
  rc auth status --profile production >/dev/null && rc products list --profile production --output json
```

#### rc auth validate

Validate authentication and API connectivity

**Examples**

```bash
# Validate the active profile
  rc auth validate

  # Validate a specific profile
  rc auth validate --profile production
```

### rc charts

View charts and metrics

#### rc charts options

Get available filter/segment options for a chart

**Examples**

```bash
# Get options for a chart
  rc charts options revenue

  # Get as JSON
  rc charts options revenue -o json
```

#### rc charts overview

Show metrics overview for a project

**Flags**

- `--interval`: refresh interval for --watch Default: `5s`.
- `-w, --watch`: continuously refresh Default: `false`.

**Examples**

```bash
# Show metrics overview
  rc charts overview

  # Watch for changes
  rc charts overview --watch

  # Get as JSON
  rc charts overview -o json
```

#### rc charts show

Show a specific chart's data

**Examples**

```bash
# Show revenue chart
  rc charts show revenue

  # Show active subscribers chart as JSON
  rc charts show active_subscribers -o json
```

### rc collaborators

View project collaborators

#### rc collaborators list

List project collaborators

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List collaborators
  rc collaborators list

  # List with JSON output
  rc collaborators list -o json
```

### rc completion

Generate shell completion scripts

Generate shell completion scripts for rc.

To load completions:

Bash:
  $ source <(rc completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ rc completion bash > /etc/bash_completion.d/rc
  # macOS:
  $ rc completion bash > $(brew --prefix)/etc/bash_completion.d/rc

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ rc completion zsh > "${fpath[1]}/_rc"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ rc completion fish | source
  # To load completions for each session, execute once:
  $ rc completion fish > ~/.config/fish/completions/rc.fish

PowerShell:
  PS> rc completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, add the output to your profile.

### rc config

Inspect rc configuration

Inspect local rc configuration, profiles, and default project settings.

#### rc config profiles

List configured profiles

**Examples**

```bash
# List profiles
  rc config profiles

  # Script-friendly profile list
  rc config profiles --output json
```

### rc currencies

Manage virtual currencies

#### rc currencies archive

Archive a virtual currency

**Examples**

```bash
# Archive a virtual currency
  rc currencies archive COINS
```

#### rc currencies balance

Show a customer's virtual currency balances

**Flags**

- `--customer-id`: customer ID (required)

**Examples**

```bash
# Check balances for a customer
  rc currencies balance --customer-id user-123
```

#### rc currencies create

Create a virtual currency

**Flags**

- `--code`: currency code, e.g. COINS (required)
- `--name`: display name (required)

**Examples**

```bash
# Create a virtual currency
  rc currencies create --code COINS --name "Gold Coins"
```

#### rc currencies credit

Create a virtual currency transaction (credit/debit)

**Flags**

- `--amount`: amount (positive=credit, negative=debit) (required) Default: `0`.
- `--code`: currency code (required)
- `--customer-id`: customer ID (required)

**Examples**

```bash
# Credit 100 coins
  rc currencies credit --customer-id user-123 --code COINS --amount 100

  # Debit 50 coins
  rc currencies credit --customer-id user-123 --code COINS --amount -50
```

#### rc currencies delete

Delete a virtual currency

**Examples**

```bash
# Delete a virtual currency
  rc currencies delete COINS
```

#### rc currencies get

Get a virtual currency by code

**Examples**

```bash
# Get currency details
  rc currencies get COINS

  # Get as JSON
  rc currencies get COINS -o json
```

#### rc currencies list

List virtual currencies

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List virtual currencies
  rc currencies list

  # List with JSON output
  rc currencies list -o json
```

#### rc currencies set-balance

Set a customer's virtual currency balance directly

**Flags**

- `--balance`: new balance value (required) Default: `0`.
- `--code`: currency code (required)
- `--customer-id`: customer ID (required)

**Examples**

```bash
# Set balance to 500
  rc currencies set-balance --customer-id user-123 --code COINS --balance 500
```

#### rc currencies unarchive

Unarchive a virtual currency

**Examples**

```bash
# Unarchive a virtual currency
  rc currencies unarchive COINS
```

#### rc currencies update

Update a virtual currency

**Flags**

- `--name`: new display name (required)

**Examples**

```bash
# Update currency name
  rc currencies update COINS --name "Premium Coins"
```

### rc customers

Manage customers and their entitlements

Look up, create, and manage RevenueCat customers.

Customers are identified by their app user ID. You can look up customer info,
check active entitlements, grant/revoke access, and manage subscriptions.

Examples:
  rc customers list
  rc customers lookup user-123
  rc customers diagnose user-123
  rc customers entitlements user-123
  rc customers subscriptions user-123
  rc customers grant --customer-id user-123 --entitlement-id entla1b2c3 --expires-at 1735689600000
  rc customers delete user-123

#### rc customers aliases

List aliases for a customer

**Examples**

```bash
# List aliases
  rc customers aliases user-123
```

#### rc customers assign-offering

Assign or clear an offering override for a customer

**Flags**

- `--clear`: clear the offering override Default: `false`.
- `--customer-id`: customer ID (required)
- `--offering-id`: offering ID to assign

**Examples**

```bash
# Assign an offering override
  rc customers assign-offering --customer-id user-123 --offering-id ofrnge1a2b3c

  # Clear the offering override
  rc customers assign-offering --customer-id user-123 --clear
```

#### rc customers attributes

List attributes for a customer

**Examples**

```bash
# List customer attributes
  rc customers attributes user-123
```

#### rc customers create

Create a customer

**Flags**

- `--id`: customer ID (required)

**Examples**

```bash
# Create a customer
  rc customers create --id user-456

  # Create and output as JSON
  rc customers create --id user-456 -o json
```

#### rc customers delete

Delete a customer

**Examples**

```bash
# Delete a customer
  rc customers delete user-123
```

#### rc customers diagnose

Diagnose why a customer does or does not have access

Diagnose why a customer does or does not have access.

The diagnosis is read-only. It looks up the customer, active entitlements,
subscriptions, purchases, aliases, and attributes, then reports likely access
issues and follow-up commands for support debugging.

**Flags**

- `--strict`: return a non-zero exit code when failed checks are found Default: `false`.

**Examples**

```bash
# Diagnose customer access
  rc customers diagnose user-123

  # Emit JSON for scripts
  rc customers diagnose user-123 --output json

  # Fail when blocking access findings are found
  rc customers diagnose user-123 --strict
```

#### rc customers entitlements

List active entitlements for a customer

**Flags**

- `--interval`: refresh interval for --watch Default: `5s`.
- `-w, --watch`: continuously refresh Default: `false`.

**Examples**

```bash
# List active entitlements
  rc customers entitlements user-123

  # List active entitlements as JSON
  rc customers entitlements user-123 --output json

  # Use a production profile
  rc customers entitlements user-123 --profile production

  # Extract entitlement IDs for scripting
  rc customers entitlements user-123 --output json | jq -r '.items[].entitlement_id'

  # Look up a customer, then list active entitlements
  rc customers lookup user-123
  rc customers entitlements user-123

  # Watch for entitlement changes
  rc customers entitlements user-123 --watch --interval 10s
```

#### rc customers grant

Grant an entitlement to a customer

**Flags**

- `--customer-id`: customer ID (required)
- `--entitlement-id`: entitlement ID to grant (required)
- `--expires-at`: expiration timestamp in ms since epoch (required) Default: `0`.

**Examples**

```bash
# Grant an entitlement with expiration
  rc customers grant --customer-id user-123 --entitlement-id entla1b2c3 --expires-at 1735689600000
```

#### rc customers invoice-file

Download an invoice file

Download an invoice file for a customer.

The file content is written to stdout. Redirect to save.

**Flags**

- `--customer-id`: customer ID (required)
- `--invoice-id`: invoice ID (required)

**Examples**

```bash
# Download an invoice to a file
  rc customers invoice-file --customer-id user-123 --invoice-id inv1a2b3c > invoice.pdf
```

#### rc customers invoices

List invoices for a customer

**Examples**

```bash
# List invoices
  rc customers invoices user-123
```

#### rc customers list

List customers in a project

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.
- `--search`: search by email (exact match)

**Examples**

```bash
# List customers
  rc customers list

  # Search by email
  rc customers list --search user@example.com

  # Fetch all pages
  rc customers list --all
```

#### rc customers lookup

Look up a customer by ID

**Flags**

- `--interval`: refresh interval for --watch Default: `5s`.
- `-w, --watch`: continuously refresh Default: `false`.

**Examples**

```bash
# Look up a customer
  rc customers lookup user-123

  # Look up a customer as JSON
  rc customers lookup user-123 --output json

  # Use a production profile
  rc customers lookup user-123 --profile production

  # Extract active entitlement IDs
  rc customers lookup user-123 --output json | jq -r '.active_entitlements.items[].entitlement_id'

  # Look up a customer, then inspect their subscriptions
  rc customers lookup user-123
  rc customers subscriptions user-123 --output json

  # Watch for changes
  rc customers lookup user-123 --watch --interval 10s
```

#### rc customers purchases

List purchases for a customer

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List customer purchases
  rc customers purchases user-123

  # Fetch all pages
  rc customers purchases user-123 --all
```

#### rc customers restore-purchase

Restore a Google Play purchase by order ID

**Flags**

- `--customer-id`: customer ID (required)
- `--order-id`: Google Play order ID (required)

**Examples**

```bash
# Restore a Google Play purchase
  rc customers restore-purchase --customer-id user-123 --order-id GPA.1234-5678-9012
```

#### rc customers revoke

Revoke a granted entitlement from a customer

**Flags**

- `--customer-id`: customer ID (required)
- `--entitlement-id`: entitlement ID to revoke (required)

**Examples**

```bash
# Revoke an entitlement
  rc customers revoke --customer-id user-123 --entitlement-id entla1b2c3
```

#### rc customers set-attributes

Set attributes on a customer

Set key=value attributes on a customer.

**Flags**

- `--attr`: attribute as key=value (required, repeatable) Default: `[]`.
- `--customer-id`: customer ID (required)

**Examples**

```bash
# Set a single attribute
  rc customers set-attributes --customer-id user-123 --attr 'plan=pro'

  # Set multiple attributes
  rc customers set-attributes --customer-id user-123 --attr '$email=user@example.com' --attr 'plan=pro'
```

#### rc customers subscriptions

List subscriptions for a customer

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List customer subscriptions
  rc customers subscriptions user-123

  # Fetch all pages
  rc customers subscriptions user-123 --all
```

#### rc customers transfer

Transfer subscriptions/purchases to another customer

**Flags**

- `--customer-id`: source customer ID (required)
- `--target-id`: target customer ID (required)

**Examples**

```bash
# Transfer purchases between customers
  rc customers transfer --customer-id user-123 --target-id user-456
```

### rc doctor

Check rc configuration, authentication, and API connectivity

**Examples**

```bash
# Check the active profile
  rc doctor

  # Check a specific profile
  rc doctor --profile staging

  # Check a specific project override
  rc doctor --project proj1a2b3c4d5
```

### rc entitlements

Manage entitlements

Manage entitlements in a RevenueCat project.

Entitlements represent levels of access that a customer can "unlock".
Products are attached to entitlements to grant access when purchased.

Examples:
  rc entitlements list
  rc entitlements get entla1b2c3d4e5
  rc entitlements create --lookup-key premium --display-name "Premium"
  rc entitlements attach --entitlement-id entla1b2c3d4e5 --product-id prod1a2b3c4d5e
  rc entitlements archive entla1b2c3d4e5

#### rc entitlements archive

Archive an entitlement

**Examples**

```bash
# Archive an entitlement
  rc entitlements archive entla1b2c3d4e5
```

#### rc entitlements attach

Attach products to an entitlement

Attach one or more products to an entitlement.

**Flags**

- `--entitlement-id`: entitlement ID (required)
- `--product-id`: product ID(s) to attach (required, comma-separated) Default: `[]`.

**Examples**

```bash
# Attach a single product
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1a2b3c

  # Attach multiple products
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1,prod2

  # Use a production profile
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1a2b3c --profile production

  # Attach the first product returned from a query
  rc products list --output json | jq -r '.items[0].id'
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1a2b3c

  # Verify attached products
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1a2b3c
  rc entitlements products entla1b2c3 --output json
```

#### rc entitlements create

Create a new entitlement

Create a new entitlement. Required flags are prompted interactively when
running in a terminal and not provided on the command line.

**Flags**

- `--display-name`: display name
- `--lookup-key`: lookup key identifier

**Examples**

```bash
# Create an entitlement
  rc entitlements create --lookup-key premium --display-name "Premium Access"

  # Create and print JSON
  rc entitlements create --lookup-key pro --display-name "Pro Access" --output json

  # Use a staging profile
  rc entitlements create --lookup-key beta --display-name "Beta Access" --profile staging

  # Capture the entitlement ID for a workflow
  rc entitlements create --lookup-key premium --display-name "Premium Access" --output json | jq -r '.id'

  # Create, then attach a product
  rc entitlements create --lookup-key premium --display-name "Premium Access"
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1a2b3c

  # Interactive mode (prompts for missing fields)
  rc entitlements create
```

#### rc entitlements delete

Delete an entitlement

**Examples**

```bash
# Delete an entitlement
  rc entitlements delete entla1b2c3d4e5
```

#### rc entitlements detach

Detach products from an entitlement

**Flags**

- `--entitlement-id`: entitlement ID (required)
- `--product-id`: product ID(s) to detach (required, comma-separated) Default: `[]`.

**Examples**

```bash
# Detach a product
  rc entitlements detach --entitlement-id entla1b2c3 --product-id prod1a2b3c

  # Detach multiple products
  rc entitlements detach --entitlement-id entla1b2c3 --product-id prod1,prod2
```

#### rc entitlements export

Export entitlements to a file (CSV or JSON)

Export entitlements to a CSV or JSON file. Format is detected from the file extension.

**Flags**

- `--file`: output file path (.csv or .json)

**Examples**

```bash
# Export to CSV
  rc entitlements export --file entitlements.csv

  # Export to JSON
  rc entitlements export --file entitlements.json
```

#### rc entitlements get

Get an entitlement by ID

**Examples**

```bash
# Get entitlement details
  rc entitlements get entla1b2c3d4e5

  # Get as JSON
  rc entitlements get entla1b2c3d4e5 -o json
```

#### rc entitlements import

Import entitlements from a file (CSV or JSON)

Import entitlements from a CSV or JSON file. Format is detected from the file extension.
Each row/entry creates a new entitlement in the project.

**Flags**

- `--file`: input file path (.csv or .json)

**Examples**

```bash
# Import from CSV
  rc entitlements import --file entitlements.csv

  # Import from JSON
  rc entitlements import --file entitlements.json
```

#### rc entitlements list

List entitlements in a project

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List all entitlements
  rc entitlements list

  # List entitlements for a specific project as JSON
  rc entitlements list --project proj1a2b3c4d5 --output json

  # Use a production profile
  rc entitlements list --profile production

  # Extract lookup keys for documentation
  rc entitlements list --output json | jq -r '.items[].lookup_key'

  # Find an entitlement, then list attached products
  rc entitlements list --output json | jq -r '.items[0].id'
  rc entitlements products entla1b2c3d4e5

  # Fetch every page
  rc entitlements list --all --limit 100
```

#### rc entitlements products

List products attached to an entitlement

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List products for an entitlement
  rc entitlements products entla1b2c3d4e5

  # Fetch all pages
  rc entitlements products entla1b2c3d4e5 --all
```

#### rc entitlements unarchive

Unarchive an entitlement

**Examples**

```bash
# Unarchive an entitlement
  rc entitlements unarchive entla1b2c3d4e5
```

#### rc entitlements update

Update an entitlement

**Flags**

- `--display-name`: new display name (required)

**Examples**

```bash
# Update display name
  rc entitlements update entla1b2c3d4e5 --display-name "Premium Plus"
```

### rc export

[beta] Export project configuration

**Stability:** `beta`

Export project configuration as a JSON file.
This includes apps, products, entitlements and their product attachments,
offerings, packages, package-product attachments, metadata, and current/archive state.

**Flags**

- `--file`: output file path (required)

**Examples**

```bash
# Export project config
  rc export --file config.json

  # Export from a specific project
  rc export --file config.json --project proj1a2b3c4d5
```

### rc import

[beta] Import project configuration from a JSON file

**Stability:** `beta`

Import a project configuration exported with 'rc export'.
Creates or reuses matching products, entitlements, offerings, and packages,
then restores product attachments, package attachments, metadata, and state.

Apps are matched to existing target apps by ID or by name and type. Use
--app-map source_app_id=target_app_id when automatic matching is not enough.

**Flags**

- `--app-map`: map source app ID to target app ID (source=target, repeatable) Default: `[]`.
- `--file`: input file path (required)

**Examples**

```bash
# Import into the default project
  rc import --file config.json

  # Import into a specific project
  rc import --file config.json --project proj_target123

  # Map a source app ID to a target app ID
  rc import --file config.json --app-map app_source=app_target
```

### rc init

Initialize rc configuration for a profile

Initialize rc configuration for a profile.

This creates ~/.rc/config.toml when needed, records the default project for
the selected profile, and can make that profile current. It does not prompt for
an API key; run rc auth login before or after init.

**Flags**

- `--current`: set the initialized profile as current Default: `true`.
- `--profile-name`: profile to initialize (defaults to active profile)

**Examples**

```bash
# Initialize the default profile with a project
  rc init --project proj1a2b3c4d5

  # Initialize a staging profile and make it current
  rc init --profile-name staging --project proj_staging --current

  # Log in and validate after initialization
  rc auth login --profile staging
  rc doctor --profile staging
```

### rc launch-check

Run a pre-launch RevenueCat readiness check

Run a pre-launch RevenueCat readiness check.

Launch check reuses the project health analyzer and summarizes whether the
active project has the minimum product, entitlement, offering, package, and
package-product paths needed before shipping.

**Flags**

- `--strict`: return a non-zero exit code when launch requirements are missing Default: `false`.

**Examples**

```bash
# Check whether the active project is ready to launch
  rc launch-check

  # Emit JSON for automation
  rc launch-check --output json

  # Fail when required launch paths are missing
  rc launch-check --strict
```

### rc mcp

[experimental] MCP server for RevenueCat operations

**Stability:** `experimental`

#### rc mcp serve

Start an MCP server over stdio

Start a Model Context Protocol server that exposes RevenueCat CLI
operations as MCP tools. The server communicates over stdin/stdout.

The API key is resolved in this order:
  1. RC_API_KEY environment variable
  2. Stored keychain / config token (same as normal CLI auth)

**Examples**

```bash
# Start the MCP server
  rc mcp serve

  # Start with an explicit API key
  RC_API_KEY=sk_test_xxx rc mcp serve
```

### rc migrate

[beta] Plan project migration workflows

**Stability:** `beta`

Plan project migration workflows.

Migration commands provide a safer UX around export/import behavior. The
project workflow currently supports dry-run planning only and never mutates the
target project.

#### rc migrate project

Dry-run a project configuration migration

Dry-run a project configuration migration.

The command exports the source project in memory, compares it with the target
project, and reports what import would create, reuse, update, attach, archive,
or skip. It requires --dry-run so migration planning stays read-only.

**Flags**

- `--app-map`: map source app ID to target app ID (source=target, repeatable) Default: `[]`.
- `--source-project`: source project ID (required)
- `--target-project`: target project ID (defaults to --project/default project)

**Examples**

```bash
# Plan a migration into the default project
  rc migrate project --source-project proj_source --dry-run

  # Plan a migration into an explicit target project
  rc migrate project --source-project proj_source --target-project proj_target --dry-run

  # Include explicit app ID mappings
  rc migrate project --source-project proj_source --target-project proj_target --app-map app_source=app_target --dry-run
```

### rc offerings

Manage offerings

Manage offerings in a RevenueCat project.

Offerings are the selection of products that are presented to a customer
on your paywall. Each offering contains one or more packages.

Examples:
  rc offerings list
  rc offerings get ofrnge1a2b3c4d5
  rc offerings create --lookup-key default --display-name "Standard Offering"
  rc offerings update ofrnge1a2b3c4d5 --is-current
  rc offerings archive ofrnge1a2b3c4d5

#### rc offerings archive

Archive an offering

**Examples**

```bash
# Archive an offering
  rc offerings archive ofrnge1a2b3c4d5
```

#### rc offerings create

Create a new offering

Create a new offering. Required flags are prompted interactively when
running in a terminal and not provided on the command line.

**Flags**

- `--display-name`: display name
- `--lookup-key`: lookup key identifier

**Examples**

```bash
# Create an offering
  rc offerings create --lookup-key default --display-name "Standard Offering"

  # Create and print JSON
  rc offerings create --lookup-key winback --display-name "Winback Offer" --output json

  # Use a staging profile
  rc offerings create --lookup-key beta --display-name "Beta Offer" --profile staging

  # Capture the offering ID
  rc offerings create --lookup-key default --display-name "Standard Offering" --output json | jq -r '.id'

  # Create an offering, then add a package
  rc offerings create --lookup-key default --display-name "Standard Offering"
  rc packages create --offering-id ofrnge1a2b3c --lookup-key monthly --display-name "Monthly"

  # Interactive mode (prompts for missing fields)
  rc offerings create
```

#### rc offerings delete

Delete an offering and its packages

**Examples**

```bash
# Delete an offering
  rc offerings delete ofrnge1a2b3c4d5
```

#### rc offerings get

Get an offering by ID

**Examples**

```bash
# Get offering details with packages
  rc offerings get ofrnge1a2b3c4d5

  # Get as JSON
  rc offerings get ofrnge1a2b3c4d5 -o json
```

#### rc offerings list

List offerings in a project

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List all offerings
  rc offerings list

  # List offerings for a specific project as JSON
  rc offerings list --project proj1a2b3c4d5 --output json

  # Use a production profile
  rc offerings list --profile production

  # Find the current offering
  rc offerings list --output json | jq -r '.items[] | select(.is_current) | .id'

  # List offerings, then inspect packages on one offering
  rc offerings list
  rc offerings get ofrnge1a2b3c4d5 --output json | jq '.packages.items'

  # Fetch every page
  rc offerings list --all --limit 100
```

#### rc offerings publish

Validate and make an offering current

Validate and make an offering current.

Publish checks that the offering is active, has packages, and that each package
has product links before setting it as the current offering.

**Examples**

```bash
# Publish an offering after validation
  rc offerings publish ofrnge1a2b3c4d5

  # Preview the checks and publish result as JSON
  rc offerings publish ofrnge1a2b3c4d5 --output json
```

#### rc offerings unarchive

Unarchive an offering

**Examples**

```bash
# Unarchive an offering
  rc offerings unarchive ofrnge1a2b3c4d5
```

#### rc offerings update

Update an offering

**Flags**

- `--display-name`: new display name
- `--is-current`: set as current offering Default: `false`.

**Examples**

```bash
# Update display name
  rc offerings update ofrnge1a2b3c4d5 --display-name "Premium Offering"

  # Set as current offering
  rc offerings update ofrnge1a2b3c4d5 --is-current
```

### rc packages

Manage packages within offerings

Manage packages within RevenueCat offerings.

Packages are the unit that ties products to offerings. Each package in an
offering can have one or more products attached (for different platforms).

Examples:
  rc packages list --offering-id ofrnge1a2b3c
  rc packages get pkge1a2b3c4d5
  rc packages create --offering-id ofrnge1a2b3c --lookup-key monthly --display-name "Monthly"
  rc packages attach --package-id pkge1a2b3c --product-id prod1a2b3c
  rc packages delete pkge1a2b3c4d5

#### rc packages attach

Attach a product to a package

Attach a product to a package with eligibility criteria.

Eligibility options: all (default), google_sdk_lt_6, google_sdk_ge_6

**Flags**

- `--eligibility`: eligibility criteria: all, google_sdk_lt_6, google_sdk_ge_6 Default: `all`.
- `--package-id`: package ID (required)
- `--product-id`: product ID (required)

**Examples**

```bash
# Attach a product to a package
  rc packages attach --package-id pkge1a2b3c --product-id prod1a2b3c

  # Attach with eligibility criteria
  rc packages attach --package-id pkge1a2b3c --product-id prod1a2b3c --eligibility google_sdk_ge_6
```

#### rc packages create

Create a new package in an offering

**Flags**

- `--display-name`: display name (required)
- `--lookup-key`: lookup key (required)
- `--offering-id`: offering ID (required)
- `--position`: display position (min 1) Default: `0`.

**Examples**

```bash
# Create a package
  rc packages create --offering-id ofrnge1a2b3c --lookup-key monthly --display-name "Monthly"

  # Create with position
  rc packages create --offering-id ofrnge1a2b3c --lookup-key annual --display-name "Annual" --position 1
```

#### rc packages delete

Delete a package

**Examples**

```bash
# Delete a package
  rc packages delete pkge1a2b3c4d5
```

#### rc packages detach

Detach products from a package

**Flags**

- `--package-id`: package ID (required)
- `--product-id`: product ID(s) to detach (required, comma-separated) Default: `[]`.

**Examples**

```bash
# Detach a product
  rc packages detach --package-id pkge1a2b3c --product-id prod1a2b3c

  # Detach multiple products
  rc packages detach --package-id pkge1a2b3c --product-id prod1,prod2
```

#### rc packages get

Get a package by ID

**Examples**

```bash
# Get package details
  rc packages get pkge1a2b3c4d5

  # Get as JSON
  rc packages get pkge1a2b3c4d5 -o json
```

#### rc packages list

List packages in an offering

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.
- `--offering-id`: offering ID (required)

**Examples**

```bash
# List packages in an offering
  rc packages list --offering-id ofrnge1a2b3c

  # List with JSON output
  rc packages list --offering-id ofrnge1a2b3c -o json

  # Fetch all pages
  rc packages list --offering-id ofrnge1a2b3c --all
```

#### rc packages products

List products attached to a package

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List products for a package
  rc packages products pkge1a2b3c4d5

  # Fetch all pages
  rc packages products pkge1a2b3c4d5 --all
```

#### rc packages update

Update a package

**Flags**

- `--display-name`: new display name
- `--position`: new position Default: `0`.

**Examples**

```bash
# Update display name
  rc packages update pkge1a2b3c4d5 --display-name "Annual Plan"

  # Update position
  rc packages update pkge1a2b3c4d5 --position 2
```

### rc paywalls

Manage paywalls

#### rc paywalls create

Create a paywall

**Examples**

```bash
# Create a paywall
  rc paywalls create
```

#### rc paywalls delete

Delete a paywall

**Examples**

```bash
# Delete a paywall
  rc paywalls delete pw1a2b3c4d5
```

#### rc paywalls get

Get a paywall by ID

**Examples**

```bash
# Get paywall details
  rc paywalls get pw1a2b3c4d5

  # Get as JSON
  rc paywalls get pw1a2b3c4d5 -o json
```

#### rc paywalls list

List paywalls

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List paywalls
  rc paywalls list

  # List with JSON output
  rc paywalls list -o json
```

#### rc paywalls validate

Validate paywall readiness

Validate paywall readiness.

The validator is read-only. It checks that paywalls exist and that the current
offering has packages with product links, which are the RevenueCat paths a
paywall needs before launch.

**Flags**

- `--strict`: return a non-zero exit code when failed checks are found Default: `false`.

**Examples**

```bash
# Validate paywall readiness
  rc paywalls validate

  # Emit JSON for automation
  rc paywalls validate --output json

  # Return non-zero when blocking checks fail
  rc paywalls validate --strict
```

### rc products

Manage products

#### rc products archive

Archive a product

**Examples**

```bash
# Archive a product
  rc products archive prod1a2b3c4d5
```

#### rc products create

Create a new product

Create a new product. Required flags are prompted interactively when
running in a terminal and not provided on the command line.

**Flags**

- `--app-id`: app ID
- `--display-name`: display name
- `--store-id`: store product identifier
- `--type`: product type: subscription, one_time, consumable, non_consumable

**Examples**

```bash
# Create a subscription product
  rc products create --store-id com.app.monthly --app-id app1a2b3c4 --type subscription

  # Create with a display name and JSON output
  rc products create --store-id com.app.yearly --app-id app1a2b3c4 --type subscription --display-name "Annual Plan" --output json

  # Use a staging profile
  rc products create --store-id com.app.monthly --app-id app1a2b3c4 --type subscription --profile staging

  # Capture the new product ID
  rc products create --store-id com.app.lifetime --app-id app1a2b3c4 --type non_consumable --output json | jq -r '.id'

  # Create, then attach to an entitlement
  rc products create --store-id com.app.yearly --app-id app1a2b3c4 --type subscription --display-name "Annual Plan"
  rc entitlements attach --entitlement-id entla1b2c3 --product-id prod1a2b3c4

  # Interactive mode (prompts for missing fields)
  rc products create
```

#### rc products delete

Delete a product

**Examples**

```bash
# Delete a product
  rc products delete prod1a2b3c4d5
```

#### rc products export

Export products to a file (CSV or JSON)

Export products to a CSV or JSON file. Format is detected from the file extension.

**Flags**

- `--file`: output file path (.csv or .json)

**Examples**

```bash
# Export to CSV
  rc products export --file products.csv

  # Export to JSON
  rc products export --file products.json
```

#### rc products get

Get a product by ID

**Examples**

```bash
# Get product details
  rc products get prod1a2b3c4d5

  # Get as JSON for scripting
  rc products get prod1a2b3c4d5 --output json

  # Use a production profile
  rc products get prod1a2b3c4d5 --profile production

  # Read the store identifier only
  rc products get prod1a2b3c4d5 --output json | jq -r '.store_identifier'

  # Find a product from the list, then inspect it
  rc products list --output json | jq -r '.items[0].id'
  rc products get prod1a2b3c4d5
```

#### rc products import

Import products from a file (CSV or JSON)

Import products from a CSV or JSON file. Format is detected from the file extension.
Each row/entry creates a new product in the project.

**Flags**

- `--file`: input file path (.csv or .json)

**Examples**

```bash
# Import from CSV
  rc products import --file products.csv

  # Import from JSON
  rc products import --file products.json
```

#### rc products list

List products in a project

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--app-id`: filter by app ID
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List all products
  rc products list

  # List products from a specific project as JSON
  rc products list --project proj1a2b3c4d5 --output json

  # Use a production profile
  rc products list --profile production

  # Extract store identifiers for a release script
  rc products list --output json | jq -r '.items[].store_identifier'

  # Review products for an app, then inspect one
  rc products list --app-id app1a2b3c4
  rc products get prod1a2b3c4d5

  # Fetch every page
  rc products list --all --limit 100
```

#### rc products push-to-store

Push a product to the store (create in the connected store)

Push a product configuration to the connected app store.

This creates the product in the store (e.g., App Store Connect, Google Play)
using the product configuration defined in RevenueCat.

**Examples**

```bash
# Push a product to App Store Connect / Google Play
  rc products push-to-store prod1a2b3c4d5
```

#### rc products unarchive

Unarchive a product

**Examples**

```bash
# Unarchive a product
  rc products unarchive prod1a2b3c4d5
```

#### rc products update

Update a product

**Flags**

- `--display-name`: new display name (required)

**Examples**

```bash
# Update display name
  rc products update prod1a2b3c4d5 --display-name "Premium Monthly"
```

### rc projects

Manage RevenueCat projects

#### rc projects create

Create a new project

**Flags**

- `--name`: project name (required)

**Examples**

```bash
# Create a new project
  rc projects create --name "My App"

  # Create and output as JSON
  rc projects create --name "My App" -o json
```

#### rc projects doctor

Check project setup for common RevenueCat launch issues

Check project setup for common RevenueCat launch issues.

The doctor reads apps, products, entitlements, offerings, packages, and
package-product links. It reports missing or incomplete relationships without
mutating the project.

**Flags**

- `--strict`: return a non-zero exit code when errors are found Default: `false`.

**Examples**

```bash
# Check the active project
  rc project doctor

  # Check a specific project and emit JSON
  rc project doctor --project proj1a2b3c4d5 --output json

  # Fail the command when project health has errors
  rc project doctor --strict
```

#### rc projects list

List all projects

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List all projects
  rc projects list

  # List with JSON output
  rc projects list -o json

  # Fetch all pages
  rc projects list --all
```

#### rc projects set-default

Set the default project for all commands

**Examples**

```bash
# Set the default project
  rc projects set-default proj1a2b3c4d5

  # Set default for a specific profile
  rc projects set-default proj1a2b3c4d5 --profile staging
```

### rc purchases

Manage one-time purchases

#### rc purchases entitlements

List entitlements for a purchase

**Examples**

```bash
# List entitlements for a purchase
  rc purchases entitlements purch1a2b3c4d5
```

#### rc purchases get

Get a purchase by ID

**Examples**

```bash
# Get purchase details
  rc purchases get purch1a2b3c4d5

  # Get as JSON
  rc purchases get purch1a2b3c4d5 -o json
```

#### rc purchases list

List purchases in a project

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List purchases
  rc purchases list

  # List with JSON output
  rc purchases list -o json

  # Fetch all pages
  rc purchases list --all
```

#### rc purchases refund

Refund a purchase

**Examples**

```bash
# Refund a purchase
  rc purchases refund purch1a2b3c4d5
```

### rc setup

Run guided setup workflows

Run guided setup workflows.

Setup commands compose lower-level RevenueCat API operations into repeatable
project configuration flows.

#### rc setup product

Set up a product access path

Set up a product access path.

This workflow creates or reuses a product, entitlement, offering, and package,
then ensures the product is attached to both the entitlement and package. It is
safe to rerun because existing resources are reused by store ID, lookup key, and
package key.

**Flags**

- `--app-id`: RevenueCat app ID (required)
- `--display-name`: product display name (defaults to --store-id)
- `--entitlement-key`: entitlement lookup key Default: `premium`.
- `--entitlement-name`: entitlement display name Default: `Premium`.
- `--make-current`: make the offering current after setup Default: `false`.
- `--offering-key`: offering lookup key Default: `default`.
- `--offering-name`: offering display name Default: `Default`.
- `--package-key`: package lookup key Default: `$rc_monthly`.
- `--package-name`: package display name Default: `Monthly`.
- `--store-id`: store product identifier (required)
- `--type`: product type: subscription, one_time, consumable, non_consumable Default: `subscription`.

**Examples**

```bash
# Set up a monthly subscription path
  rc setup product \
    --app-id app1a2b3c4 \
    --store-id com.example.app.monthly \
    --display-name "Monthly" \
    --entitlement-key premium \
    --offering-key default \
    --package-key '$rc_monthly'

  # Also make the offering current
  rc setup product --app-id app1a2b3c4 --store-id com.example.app.monthly --make-current
```

### rc subscriptions

Manage subscriptions

View and manage RevenueCat subscriptions.

Examples:
  rc subscriptions list
  rc subscriptions get sub1ab2c3d4e5
  rc subscriptions transactions sub1ab2c3d4e5
  rc subscriptions cancel sub1ab2c3d4e5
  rc subscriptions refund sub1ab2c3d4e5

#### rc subscriptions cancel

Cancel a subscription

**Examples**

```bash
# Cancel a subscription
  rc subscriptions cancel sub1ab2c3d4e5
```

#### rc subscriptions entitlements

List entitlements for a subscription

**Examples**

```bash
# List entitlements for a subscription
  rc subscriptions entitlements sub1ab2c3d4e5
```

#### rc subscriptions get

Get a subscription by ID

**Flags**

- `--interval`: refresh interval for --watch Default: `5s`.
- `-w, --watch`: continuously refresh Default: `false`.

**Examples**

```bash
# Get subscription details
  rc subscriptions get sub1ab2c3d4e5

  # Get as JSON
  rc subscriptions get sub1ab2c3d4e5 --output json

  # Use a production profile
  rc subscriptions get sub1ab2c3d4e5 --profile production

  # Extract the authenticated management URL
  rc subscriptions management-url sub1ab2c3d4e5 --output json | jq -r '.url'

  # Inspect a subscription, then list transactions
  rc subscriptions get sub1ab2c3d4e5
  rc subscriptions transactions sub1ab2c3d4e5 --output json

  # Watch for changes
  rc subscriptions get sub1ab2c3d4e5 --watch --interval 10s
```

#### rc subscriptions list

List subscriptions in a project

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List subscriptions
  rc subscriptions list

  # List subscriptions for a specific project as JSON
  rc subscriptions list --project proj1a2b3c4d5 --output json

  # Use a production profile
  rc subscriptions list --profile production

  # Extract active subscription IDs
  rc subscriptions list --output json | jq -r '.items[] | select(.status == "active") | .id'

  # Find a subscription, then inspect transactions
  rc subscriptions list --output json | jq -r '.items[0].id'
  rc subscriptions transactions sub1ab2c3d4e5

  # Fetch every page
  rc subscriptions list --all --limit 100
```

#### rc subscriptions management-url

Get authenticated management URL for a subscription

**Examples**

```bash
# Get management URL
  rc subscriptions management-url sub1ab2c3d4e5
```

#### rc subscriptions refund

Refund a subscription

**Examples**

```bash
# Refund a subscription
  rc subscriptions refund sub1ab2c3d4e5
```

#### rc subscriptions refund-transaction

Refund a specific transaction within a subscription

**Flags**

- `--transaction-id`: transaction ID to refund (required)

**Examples**

```bash
# Refund a specific transaction
  rc subscriptions refund-transaction sub1ab2c3d4e5 --transaction-id txn1a2b3c
```

#### rc subscriptions transactions

List transactions for a subscription

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List transactions
  rc subscriptions transactions sub1ab2c3d4e5

  # Fetch all pages
  rc subscriptions transactions sub1ab2c3d4e5 --all
```

### rc version

Print the version of rc

### rc webhooks

Manage webhook integrations

#### rc webhooks create

Create a new webhook integration

**Flags**

- `--name`: webhook name (required)
- `--url`: webhook endpoint URL (required)

**Examples**

```bash
# Create a webhook
  rc webhooks create --name "My Webhook" --url https://example.com/webhook
```

#### rc webhooks delete

Delete a webhook integration

**Examples**

```bash
# Delete a webhook
  rc webhooks delete wh1a2b3c4d5
```

#### rc webhooks get

Get a webhook by ID

**Examples**

```bash
# Get webhook details
  rc webhooks get wh1a2b3c4d5
```

#### rc webhooks list

List webhook integrations

**Flags**

- `--all`: fetch all pages Default: `false`.
- `--limit`: max items per page Default: `0`.

**Examples**

```bash
# List webhooks
  rc webhooks list

  # List with JSON output
  rc webhooks list -o json
```

#### rc webhooks update

Update a webhook integration

**Flags**

- `--name`: new webhook name
- `--url`: new webhook URL

**Examples**

```bash
# Update webhook name
  rc webhooks update wh1a2b3c4d5 --name "Updated Webhook"

  # Update webhook URL
  rc webhooks update wh1a2b3c4d5 --url https://example.com/new-webhook
```

### rc whoami

Show the active profile, auth source, and default project

**Examples**

```bash
# Show current identity context
  rc whoami

  # Script-friendly output
  rc whoami --output json
```

