# RevenueCat API Notes

Implementation notes captured from the CLI and test harness.

## Request Semantics

- RevenueCat API v2 update operations use `POST`, not `PUT` or `PATCH`.
- Mutating commands send JSON bodies through `internal/api.Client.Post`.
- Archive and unarchive operations use action endpoints such as
  `/actions/archive` and `/actions/unarchive`.
- Attach and detach operations use action endpoints with explicit ID arrays.

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
- Always run import with `--dry-run` before applying a project migration.

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
