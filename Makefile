BINARY  := gogoversion
LINK    := ggv
INSTALL := $(shell go env GOPATH)/bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.DEFAULT_GOAL := help

.PHONY: help build install uninstall clean run tidy

help: ## Show available targets
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / && $$1 != "help" {printf "%s\t%s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort | awk -F '\t' '{printf "  %-10s %s\n", $$1, $$2}'

build: ## Build binary
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) .

tidy: ## Tidy Go modules
	go mod tidy

install: build ## Install binary and symlink
	cp $(BINARY) $(INSTALL)/$(BINARY)
	ln -sf $(INSTALL)/$(BINARY) $(INSTALL)/$(LINK)

uninstall: ## Remove installed binary and symlink
	rm -f $(INSTALL)/$(BINARY) $(INSTALL)/$(LINK)

clean: ## Remove local binary
	rm -f $(BINARY)

run: ## Run with --dry-run
	go run . --dry-run .