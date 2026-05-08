.PHONY: all build test vet tools lint analyze check vulncheck

GOBIN_DIR := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)
GOLANGCI_LINT := $(GOBIN_DIR)/golangci-lint

all: build

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

tools:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

lint: tools
	$(GOLANGCI_LINT) run

vulncheck: tools
	$(GOBIN_DIR)/govulncheck ./...

analyze: vet lint

check: build vet test

build:
	go build ./...
