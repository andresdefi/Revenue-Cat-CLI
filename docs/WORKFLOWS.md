# rc Workflows

Copyable workflows for common RevenueCat operations. Commands favor explicit IDs
and JSON output for repeatable scripts.

## First Run

```bash
rc auth login
rc projects list --output table
rc projects set-default proj1a2b3c4d5
rc doctor
rc project doctor
rc launch-check
rc products list
```

## Check Project Readiness

```bash
rc project doctor --output table
rc project doctor --output json
rc project doctor --strict
```

`rc project doctor` checks apps, products, entitlements, offerings, packages,
and package-product links. Use `--strict` in CI when a failed project health
check should stop the run.

```bash
rc launch-check --output table
rc launch-check --strict
```

`rc launch-check` summarizes whether the project has the required launch paths:
an app, active products, active entitlements with product links, a current
offering, packages, and package-product links.

## Configure App Store Credentials

```bash
rc apps update app1a2b3c4d5 \
  --shared-secret 1234567890abcdef1234567890abcdef
```

```bash
rc apps update app1a2b3c4d5 \
  --subscription-key-file ./SubscriptionKey_ABC123.p8 \
  --subscription-key-id ABC123 \
  --subscription-key-issuer 5a049d62-1b9b-453c-b605-1988189d8129
```

`--subscription-key-file` reads the `.p8` file from disk and sends its contents
as `app_store.subscription_private_key`. RevenueCat API v2 does not currently
document a Google Play service-account credential field for app updates, so
`--service-account-file` fails with guidance instead of sending an unsupported
payload.

## Set Up Monthly And Yearly Products

```bash
rc apps list --output table

rc setup product \
  --app-id app1a2b3c4d5 \
  --store-id com.example.app.monthly \
  --display-name "Monthly" \
  --entitlement-key premium \
  --offering-key default \
  --package-key '$rc_monthly'
```

`rc setup product` creates or reuses the product, entitlement, offering,
package, and product attachments needed for a working access path. Use the
lower-level commands below when you need separate review or custom sequencing.

```bash
rc products create \
  --app-id app1a2b3c4d5 \
  --store-id com.example.app.monthly \
  --type subscription \
  --display-name "Monthly"

rc products create \
  --app-id app1a2b3c4d5 \
  --store-id com.example.app.yearly \
  --type subscription \
  --display-name "Yearly"
```

## Attach Products To An Entitlement

```bash
rc entitlements create --lookup-key premium --display-name "Premium"

rc entitlements attach \
  --entitlement-id entl1a2b3c4d5 \
  --product-id prod_monthly,prod_yearly

rc entitlements products entl1a2b3c4d5 --all
```

## Create A Default Offering

```bash
rc offerings create --lookup-key default --display-name "Default"

rc packages create \
  --offering-id ofrnge1a2b3c4d5 \
  --lookup-key '$rc_monthly' \
  --display-name "Monthly"

rc packages attach \
  --package-id pkge1a2b3c4d5 \
  --product-id prod_monthly

rc offerings publish ofrnge1a2b3c4d5
```

`rc offerings publish` validates that packages and product links exist before
making the offering current.

## Validate Paywalls

```bash
rc paywalls validate --output table
rc paywalls validate --strict
```

`rc paywalls validate` checks that paywalls exist and that the current offering
has packages with package-product links.

## Inspect Customer Access

```bash
rc customers diagnose user_123 --output table
rc customers lookup user_123 --output json --pretty
rc customers entitlements user_123 --output table
rc customers subscriptions user_123 --all --output table
rc customers purchases user_123 --all --output table
```

Start with `rc customers diagnose` for a read-only support summary. It reports
whether active entitlement access is present, lists subscriptions and purchases,
surfaces aliases that may explain split identity, and suggests follow-up CLI
commands. Use `--strict` when failed access checks should return a non-zero exit
code in automation.

If a customer does not have expected access, check the product-to-entitlement
links first, then inspect the customer purchases and subscriptions.

## Move Project Configuration Safely

```bash
rc migrate project \
  --source-project proj_source \
  --target-project proj_target \
  --app-map app_source_ios=app_target_ios \
  --dry-run

rc export --project proj_source --file project-config.json

rc import \
  --project proj_target \
  --file project-config.json \
  --app-map app_source_ios=app_target_ios \
  --dry-run

rc import \
  --project proj_target \
  --file project-config.json \
  --app-map app_source_ios=app_target_ios
```

`rc migrate project --dry-run`, `rc export`, and `rc import` are beta. Start
with the migration dry-run plan to inspect creates, reuses, skips, and required
app mappings before applying an export/import migration.
