.PHONY: all build test vet tools lint analyze check vulncheck

GOBIN_DIR := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)
GOLANGCI_LINT := $(GOBIN_DIR)/golangci-lint
GOVULNCHECK   := $(GOBIN_DIR)/govulncheck

# Pinned tool versions. Bump deliberately; @latest is fragile in CI.
GOLANGCI_LINT_VERSION ?= v2.6.0
GOVULNCHECK_VERSION   ?= v1.1.4

all: build

build:
	go build ./...

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

lint: tools
	$(GOLANGCI_LINT) run

vulncheck: tools
	$(GOVULNCHECK) ./...

# `analyze` is the static-only flow that mirrors go-gd's CI.
# golangci-lint already runs govet, so we don't invoke it twice here.
analyze: lint

# `check` is the full local verification flow used as the CI gate.
check: build analyze test vulncheck
