# rc Workflows

Copyable workflows for common RevenueCat operations. Commands favor explicit IDs
and JSON output for repeatable scripts.

## First Run

```bash
rc auth login
rc projects list --output table
rc projects set-default proj1a2b3c4d5
rc doctor
rc products list
```

## Set Up Monthly And Yearly Products

```bash
rc apps list --output table

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

rc offerings update ofrnge1a2b3c4d5 --is-current
```

## Inspect Customer Access

```bash
rc customers lookup user_123 --output json --pretty
rc customers entitlements user_123 --output table
rc customers subscriptions user_123 --all --output table
rc customers purchases user_123 --all --output table
```

If a customer does not have expected access, check the product-to-entitlement
links first, then inspect the customer purchases and subscriptions.

## Move Project Configuration Safely

```bash
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

`rc export` and `rc import` are beta. Use `--dry-run` first and inspect warnings
for missing app or product mappings before applying changes.
