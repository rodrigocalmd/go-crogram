BIN     := crogram
PKG     := ./cmd/crogram
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

# Where "go install" puts binaries: GOBIN, or GOPATH/bin when GOBIN is empty.
GOBIN_DIR := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)

.PHONY: help build install uninstall test lint examples

help: ## show this help
	@grep -E '^[a-z]+:.*##' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-10s %s\n", $$1, $$2}'

build: ## build ./crogram in the current directory
	go build -ldflags "-X main.version=$(VERSION)" -o $(BIN) $(PKG)

install: ## install crogram into GOBIN (or GOPATH/bin)
	go install -ldflags "-X main.version=$(VERSION)" $(PKG)
	@echo "installed $(GOBIN_DIR)/$(BIN)"

uninstall: ## remove crogram from GOBIN (or GOPATH/bin)
	@if [ -f "$(GOBIN_DIR)/$(BIN)" ]; then \
		rm -f "$(GOBIN_DIR)/$(BIN)" && echo "removed $(GOBIN_DIR)/$(BIN)"; \
	else \
		echo "$(GOBIN_DIR)/$(BIN) not found, nothing to remove"; \
	fi

test: ## run vet and the tests with the race detector
	go vet ./...
	go test -race -cover ./...

lint: ## fail if any file is not gofmt-formatted
	@test -z "$$(gofmt -l $$(git ls-files '*.go') 2>/dev/null)" || { gofmt -l .; exit 1; }

examples: ## run the CLI examples
	sh examples/cli.sh
