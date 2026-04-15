BINARY := rc
MODULE := github.com/andresdefi/rc
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X $(MODULE)/internal/version.Version=$(VERSION) -X $(MODULE)/internal/version.Commit=$(COMMIT) -X $(MODULE)/internal/version.Date=$(DATE)

.PHONY: build test test-integration lint vet fmt docs clean install check tools security help

## build: Build the rc binary
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## install: Install rc to $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" .

## test: Run all tests
test:
	go test -race -coverprofile=coverage.out ./...

## test-integration: Run integration tests (requires RC_INTEGRATION_KEY)
test-integration:
	go test -race -tags integration ./...

## docs: Generate command reference docs
docs:
	go run ./scripts/generate-commands-docs

## lint: Run golangci-lint
lint:
	golangci-lint run

## vet: Run go vet
vet:
	go vet ./...

## fmt: Format code with gofumpt
fmt:
	gofumpt -w .

## security: Run gosec security scanner
security:
	gosec -exclude=G304,G101,G115 ./...

## check: Run all checks (fmt, docs, vet, lint, test)
check: fmt docs vet lint test

## tools: Install dev dependencies
tools:
	go install mvdan.cc/gofumpt@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

## clean: Remove build artifacts
clean:
	rm -f $(BINARY) coverage.out

## coverage: Show test coverage in browser
coverage: test
	go tool cover -html=coverage.out

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':'
