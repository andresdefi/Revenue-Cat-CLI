BINARY := rc
MODULE := github.com/andresdefi/rc
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X $(MODULE)/internal/version.Version=$(VERSION) -X $(MODULE)/internal/version.Commit=$(COMMIT) -X $(MODULE)/internal/version.Date=$(DATE)

.PHONY: build test lint vet fmt clean install check help

## build: Build the rc binary
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## install: Install rc to $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" .

## test: Run all tests
test:
	go test -race -coverprofile=coverage.out ./...

## lint: Run golangci-lint
lint:
	golangci-lint run

## vet: Run go vet
vet:
	go vet ./...

## fmt: Format code
fmt:
	gofmt -s -w .

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test

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
