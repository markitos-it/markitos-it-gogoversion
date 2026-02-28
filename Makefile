BINARY  := gogoversion
LINK    := ggv
INSTALL := $(shell go env GOPATH)/bin

.DEFAULT_GOAL := help

.PHONY: help build install uninstall clean run tidy

help:
	@echo "Available targets:"
	@echo "  build      Build binary"
	@echo "  tidy       Tidy Go modules"
	@echo "  install    Install binary and symlink"
	@echo "  uninstall  Remove installed binary and symlink"
	@echo "  clean      Remove local binary"
	@echo "  run        Run with --dry-run"

build:
	go build -o $(BINARY) .

tidy:
	go mod tidy

install: build
	cp $(BINARY) $(INSTALL)/$(BINARY)
	ln -sf $(INSTALL)/$(BINARY) $(INSTALL)/$(LINK)

uninstall:
	rm -f $(INSTALL)/$(BINARY) $(INSTALL)/$(LINK)

clean:
	rm -f $(BINARY)

run:
	go run . --dry-run