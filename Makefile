# zen-sdk Makefile

.PHONY: help test lint fmt vet build clean all

# Colors
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

## help: Display this help message
help:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "zen-sdk Makefile"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'
	@echo ""

## test: Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✅ Tests complete$(NC)"

## lint: Run linters
lint: fmt vet

## fmt: Run go fmt
fmt:
	@echo "$(GREEN)Running go fmt...$(NC)"
	@UNFORMATTED=$$(gofmt -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
		echo "$(RED)❌ Code not formatted:$(NC)"; \
		echo "$$UNFORMATTED"; \
		echo "$(YELLOW)Run: gofmt -w .$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✅ Code formatted$(NC)"

## vet: Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✅ go vet passed$(NC)"

## tidy: Tidy go.mod
tidy:
	@echo "$(GREEN)Tidying go.mod...$(NC)"
	go mod tidy
	@echo "$(GREEN)✅ go.mod tidied$(NC)"

## clean: Clean build artifacts
clean:
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	rm -f coverage.out
	go clean -cache -testcache
	@echo "$(GREEN)✅ Clean complete$(NC)"

## all: Run all checks (lint, test)
all: lint test
	@echo ""
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"
	@echo "$(GREEN)✅ All checks passed!$(NC)"
	@echo "$(GREEN)━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━$(NC)"

.DEFAULT_GOAL := help


check:
	@scripts/ci/check.sh

test-race:
	@go test -v -race -timeout=15m ./...

## zenctl: Build zenctl CLI
zenctl:
	@echo "$(GREEN)Building zenctl...$(NC)"
	@go build -ldflags "$(GO_LDFLAGS)" -o zenctl ./cmd/zenctl
	@echo "$(GREEN)✅ zenctl built$(NC)"

## install-zenctl: Install zenctl to GOBIN
install-zenctl:
	@echo "$(GREEN)Installing zenctl...$(NC)"
	@go install -ldflags "$(GO_LDFLAGS)" ./cmd/zenctl
	@GOBIN=$$(go env GOBIN); \
	if [ -z "$$GOBIN" ]; then \
		GOBIN="$$(go env GOPATH)/bin"; \
	fi; \
	echo "$(GREEN)✅ zenctl installed to $$GOBIN$(NC)"

## release-zenctl: Build zenctl for multiple architectures
release-zenctl:
	@echo "$(GREEN)Building zenctl releases...$(NC)"
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(GO_LDFLAGS)" -o dist/zenctl-linux-amd64 ./cmd/zenctl
	@GOOS=linux GOARCH=arm64 go build -ldflags "$(GO_LDFLAGS)" -o dist/zenctl-linux-arm64 ./cmd/zenctl
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(GO_LDFLAGS)" -o dist/zenctl-darwin-amd64 ./cmd/zenctl
	@GOOS=darwin GOARCH=arm64 go build -ldflags "$(GO_LDFLAGS)" -o dist/zenctl-darwin-arm64 ./cmd/zenctl
	@echo "$(GREEN)✅ zenctl releases built in dist/$(NC)"

# Version and build info for ldflags
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_LDFLAGS = -X github.com/kube-zen/zen-sdk/cmd/zenctl/internal/version.Version=$(VERSION) \
             -X github.com/kube-zen/zen-sdk/cmd/zenctl/internal/version.GitCommit=$(GIT_COMMIT) \
             -X github.com/kube-zen/zen-sdk/cmd/zenctl/internal/version.BuildTime=$(BUILD_TIME)
