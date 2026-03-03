BINARY  := gogoversion
LINK    := ggv
INSTALL := $(shell go env GOPATH)/bin
APP_PKG := ./cmd/app
REMOTE_BINARY := markitos-it-gogoversion
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.DEFAULT_GOAL := help

.PHONY: help build install uninstall clean clean-cache run-dry tidy test test-v cover run-real run-no-tag run-no-changelog

help: ## Show available targets
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / && $$1 != "help" {printf "%s\t%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort | awk -F '\t' '{printf "  %-10s %s\n", $$1, $$2}'

build: ## Build binary
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) $(APP_PKG)

tidy: ## Tidy Go modules
	go mod tidy

install: build ## Install binary and symlink
	cp $(BINARY) $(INSTALL)/$(BINARY)
	ln -sf $(INSTALL)/$(BINARY) $(INSTALL)/$(LINK)

uninstall: ## Remove installed binaries and symlink
	rm -f $(INSTALL)/$(BINARY) $(INSTALL)/$(REMOTE_BINARY) $(INSTALL)/$(LINK)

clean: ## Remove local binary
	rm -f $(BINARY)

clean-cache: ## Remove Go build cache
	go clean -cache -testcache -modcache

run-dry: ## Run with --dry-run
	go run $(APP_PKG) --dry-run .

run-real: ## Run with current directory
	go run $(APP_PKG) .

run-no-tag: ## Run with --no-tag
	go run $(APP_PKG) --no-tag .

run-no-changelog: ## Run with --no-changelog
	go run $(APP_PKG) --no-changelog .

test: ## Run tests
	go test ./...

test-v: ## Run tests verbose
	go test ./... -v

cover: ## Run tests with coverage
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out