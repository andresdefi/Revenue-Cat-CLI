# RevenueCat API Notes

Implementation notes captured from the CLI and test harness.

## API Semantics Audit - 2026-04-17

Compared the CLI against the official RevenueCat API v2 Redoc/OpenAPI
specification downloaded from
`https://www.revenuecat.com/docs/redocusaurus/plugin-redoc-0.yaml`.

Confirmed and fixed in the CLI:

- `POST /projects/{project_id}/paywalls` requires `offering_id`. `rc paywalls
  create` now requires `--offering-id` and sends `{ "offering_id": "..." }`.
- Top-level `GET /projects/{project_id}/subscriptions` is a search endpoint,
  not a broad project list. It requires `store_subscription_identifier`; `rc
  subscriptions list` now requires `--store-subscription-id`.
- Top-level `GET /projects/{project_id}/purchases` is a search endpoint, not a
  broad project list. It requires `store_purchase_identifier`; `rc purchases
  list` now requires `--store-purchase-id`.
- Customer virtual-currency transaction and update-balance endpoints require an
  `adjustments` object. `rc currencies credit` and `rc currencies set-balance`
  now send `{ "adjustments": { "<code>": <amount> } }`.
- The authenticated subscription management URL response uses
  `management_url`, not `url`.
- `POST /offerings/{offering_id}/actions/unarchive` accepts optional
  `unarchive_referenced_entities`. `rc offerings unarchive` now exposes
  `--unarchive-referenced-entities`.
- App Store app updates document App Store Connect API-key fields under
  `app_store`; `rc apps update` now exposes the documented
  `--app-store-connect-*` flags and shorter `--asc-*` aliases.
- `POST /products/{product_id}/create_in_store` allows an omitted body for
  in-app purchase products, but subscription products require
  `store_information`. `rc products push-to-store` now keeps the no-body path
  for in-app purchases and exposes subscription store-information flags for
  subscriptions.

## Request Semantics

- RevenueCat API v2 update operations use `POST`, not `PUT` or `PATCH`.
- Mutating commands send JSON bodies through `internal/api.Client.Post`.
- Archive and unarchive operations use action endpoints such as
  `/actions/archive` and `/actions/unarchive`.
- Product push-to-store can send no body for in-app purchase products, or
  `store_information.duration`, `store_information.subscription_group_name`,
  and optionally `store_information.subscription_group_id` for subscription
  products.
- Attach and detach operations use action endpoints with explicit ID arrays.
- Offering unarchive can optionally cascade to archived referenced products via
  `unarchive_referenced_entities`.
- App updates accept nested store-specific objects. `rc apps update` sends App
  Store credential fields under `app_store`, including `shared_secret`,
  `subscription_private_key`, `subscription_key_id`, and
  `subscription_key_issuer`, plus App Store Connect API-key fields
  `app_store_connect_api_key`, `app_store_connect_api_key_id`,
  `app_store_connect_api_key_issuer`, and
  `app_store_connect_vendor_number`.
- The current RevenueCat v2 OpenAPI describes `subscription_private_key` as the
  `.p8` PEM file contents. The CLI reads `--subscription-key-file` and sends the
  file contents as-is.
- `rc apps creds status` reports credential configuration booleans returned by
  RevenueCat, such as `subscription_key_configured` and
  `app_store_connect_api_key_configured`. It does not verify credential validity
  with Apple.
- The current RevenueCat v2 OpenAPI does not document a Google Play service
  account credential field on app updates, and `play_store` disallows extra
  fields. `--service-account-file` returns a clear error instead of sending an
  undocumented credential payload.

## Pagination

- List responses use `{ "object": "list", "items": [...], "next_page": ... }`.
- `rc --all` commands follow `next_page` until it is null.
- The client accepts absolute next-page URLs, `/v2/...` paths, and bare resource
  paths. This is covered by API unit tests and command-level pagination tests.

## Project Transfer

- `rc export`/`rc import` is beta.
- The export schema version is `2`.
- Export includes apps, products, entitlement product attachments, offerings,
  packages, package-product attachments, metadata, current offering state, and
  archive state where the API exposes it.
- Import maps source apps to target apps by ID, then by name/type, then by
  explicit `--app-map source=target` overrides.
- Import warns and skips dependent links when a source product cannot be mapped
  into the target project.
- Fixture tests cover archived state export/import, idempotent re-imports,
  offering/package update bodies, partial product create failures, failed
  attachment/archive calls, attachment restoration, and migration dry-run
  archive planning.
- Always run `rc migrate project --dry-run` before applying an export/import
  migration.

## Error Handling

- API error envelopes are parsed into `api.Error` with type, message, param,
  doc URL, retryable flag, and status code.
- Retryable API errors are retried with backoff.
- Non-JSON error responses are surfaced as plain HTTP errors.

## Test Coverage Expectations

- Add golden request-body tests for every new mutating command.
- Add pagination coverage for every new `--all` list command.
- Add integration tests behind the `integration` build tag for real API
  behavior that cannot be proven with fixtures.

## Project Health

- `rc project doctor` is read-only.
- It fetches apps, products, entitlements, entitlement-product links, offerings,
  packages, and package-product links with `PaginateAll`.
- The command reports failed setup checks by default without mutating anything.
  Use `--strict` when failed checks should produce a non-zero exit code.
- `rc launch-check` reuses the project health analyzer and adds launch-readiness
  checks for the required product, entitlement, current offering, package, and
  package-product paths.
