# Testing

## Local Checks

```bash
go test ./...
env PATH="$HOME/go/bin:$PATH" make check
```

`make check` runs gofumpt, go vet, golangci-lint, and race tests with coverage.

## Command Fixture Tests

Command tests use `internal/cmdtest`, which starts a local HTTP server and points
the API client at it. Use these tests for command behavior, output shape, request
paths, and request bodies.

For mutating commands, add a golden request-body assertion. Keep coverage
systematic across create/update/archive/unarchive/attach/detach/import paths,
and add fixture behavior tests when the important contract is retry behavior,
API error surfacing, archived-resource handling, or import idempotency:

```go
cmdtest.AssertRequestJSON(t, result, "POST", "/projects/proj/products", map[string]any{
    "store_identifier": "com.example.product",
    "app_id": "app1",
    "type": "subscription",
})
```

For `--all` list commands, add or update pagination contract coverage in
`cmd/pagination_contract_test.go`.

## Integration Tests

Integration tests are opt-in and use the `integration` build tag.

```bash
RC_INTEGRATION_KEY=sk_live_or_sandbox \
RC_INTEGRATION_PROJECT_ID=proj1a2b3c4d5 \
go test -race -tags integration ./internal/integration
```

Keep integration tests low-risk by default. Prefer read-only tests unless the
test creates resources with unique names and cleans them up.
