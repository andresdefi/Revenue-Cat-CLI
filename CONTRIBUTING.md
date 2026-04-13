# Contributing to rc

Thanks for your interest in contributing! This project is open to everyone.

## Getting started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/rc.git`
3. Install Go 1.22+
4. Run `go mod tidy` to install dependencies
5. Build: `go build -o rc .`

## Making changes

1. Create a branch: `git checkout -b feature/your-feature`
2. Make your changes
3. Run tests: `go test ./...`
4. Build and verify: `go build -o rc . && ./rc --help`
5. Commit and push
6. Open a pull request

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use `cobra.Command` for new commands
- All API calls go through `internal/api/client.go`
- New resource types go in `internal/api/types.go`
- Keep error messages actionable (tell the user what to do next)

## Adding a new command group

1. Create `cmd/<resource>/<resource>.go`
2. Define `New<Resource>Cmd()` that returns a `*cobra.Command`
3. Add it to `cmd/root.go`
4. Add API types to `internal/api/types.go`
5. Update README.md with the new commands

## Reporting issues

Please include:
- `rc` version (`rc --version` once versioning is added)
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
