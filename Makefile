# Makefile for hook-vault-radar
# A hook framework integration for Vault Radar scanning

# Binary name
BINARY_NAME := hook-vault-radar

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Install path
INSTALL_PATH ?= $(HOME)/.agent-hooks/vault-radar

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)"

# Directories
SRC_DIRS := ./cmd ./internal ./pkg
TEST_DIRS := ./...

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m

# Default target
.DEFAULT_GOAL := help

# Phony targets
.PHONY: all build install clean test test-integration fmt vet lint mod-tidy mod-verify run-test help

## all: Build the binary
all: build

## build: Compile the binary with version information
build:
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Building $(BINARY_NAME) $(VERSION)...$(COLOR_RESET)"
	@$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v
	@echo "$(COLOR_GREEN)✓ Build complete: $(BINARY_NAME)$(COLOR_RESET)"

## install: Install the binary to $(INSTALL_PATH)
install: build
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Installing $(BINARY_NAME) to $(INSTALL_PATH)...$(COLOR_RESET)"
	@mkdir -p $(INSTALL_PATH)
	@cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@chmod +x $(INSTALL_PATH)/$(BINARY_NAME)
	@echo "$(COLOR_GREEN)✓ Installed to $(INSTALL_PATH)/$(BINARY_NAME)$(COLOR_RESET)"

## clean: Remove build artifacts
clean:
	@echo "$(COLOR_YELLOW)Cleaning build artifacts...$(COLOR_RESET)"
	@$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@echo "$(COLOR_GREEN)✓ Clean complete$(COLOR_RESET)"

## test: Run all tests with coverage
test:
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Running tests...$(COLOR_RESET)"
	@$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic $(TEST_DIRS)
	@echo "$(COLOR_GREEN)✓ Tests complete$(COLOR_RESET)"

## test-coverage: Run tests and display coverage report
test-coverage: test
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Coverage report:$(COLOR_RESET)"
	@$(GOCMD) tool cover -func=coverage.out

## test-integration: Test with sample fixtures from testdata
test-integration: build
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Running integration tests...$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Testing with clean input:$(COLOR_RESET)"
	@cat testdata/claude/userpromptsubmit_clean.json | ./$(BINARY_NAME) --framework claude --log-level debug 2>&1
	@echo "\n$(COLOR_YELLOW)Testing with secret-containing input:$(COLOR_RESET)"
	@cat testdata/claude/userpromptsubmit.json | ./$(BINARY_NAME) --framework claude --log-level debug 2>&1 || true
	@echo "$(COLOR_GREEN)✓ Integration tests complete$(COLOR_RESET)"

## fmt: Format Go code
fmt:
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Formatting code...$(COLOR_RESET)"
	@$(GOFMT) ./...
	@echo "$(COLOR_GREEN)✓ Format complete$(COLOR_RESET)"

## vet: Run go vet for static analysis
vet:
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Running go vet...$(COLOR_RESET)"
	@$(GOVET) $(TEST_DIRS)
	@echo "$(COLOR_GREEN)✓ Vet complete$(COLOR_RESET)"

## lint: Run golangci-lint (if available)
lint:
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Running linter...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(COLOR_GREEN)✓ Lint complete$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not found, skipping lint$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)  Install with: brew install golangci-lint$(COLOR_RESET)"; \
	fi

## mod-tidy: Tidy and verify go modules
mod-tidy:
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Tidying modules...$(COLOR_RESET)"
	@$(GOMOD) tidy
	@$(GOMOD) verify
	@echo "$(COLOR_GREEN)✓ Modules tidied$(COLOR_RESET)"

## mod-verify: Verify go modules
mod-verify:
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Verifying modules...$(COLOR_RESET)"
	@$(GOMOD) verify
	@echo "$(COLOR_GREEN)✓ Modules verified$(COLOR_RESET)"

## run-test: Quick test with sample input (clean)
run-test: build
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Testing with clean sample input...$(COLOR_RESET)"
	@cat testdata/claude/userpromptsubmit_clean.json | ./$(BINARY_NAME) --framework claude --log-level info

## run-test-secret: Quick test with secret-containing input
run-test-secret: build
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Testing with secret-containing input...$(COLOR_RESET)"
	@cat testdata/claude/userpromptsubmit.json | ./$(BINARY_NAME) --framework claude --log-level info || true

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "$(COLOR_GREEN)$(COLOR_BOLD)✓ All checks passed$(COLOR_RESET)"

## release: Build release binaries for multiple platforms
release:
	@echo "$(COLOR_BLUE)$(COLOR_BOLD)Building release binaries...$(COLOR_RESET)"
	@mkdir -p dist
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe
	@GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-arm64.exe
	@echo "$(COLOR_GREEN)✓ Release binaries built in dist/$(COLOR_RESET)"

## version: Display version information
version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(shell $(GOCMD) version)"

## help: Show this help message
help:
	@echo "$(COLOR_BOLD)$(BINARY_NAME) Makefile$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Usage:$(COLOR_RESET)"
	@echo "  make [target]"
	@echo ""
	@echo "$(COLOR_BOLD)Available targets:$(COLOR_RESET)"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'
	@echo ""
	@echo "$(COLOR_BOLD)Examples:$(COLOR_RESET)"
	@echo "  make build              # Build the binary"
	@echo "  make install            # Build and install to ~/.local/bin"
	@echo "  make test               # Run all tests"
	@echo "  make check              # Run all checks (fmt, vet, lint, test)"
	@echo "  make run-test           # Quick test with sample input"
	@echo "  make clean              # Remove build artifacts"
	@echo ""
	@echo "$(COLOR_BOLD)Configuration:$(COLOR_RESET)"
	@echo "  INSTALL_PATH            Installation directory (default: ~/.local/bin)"
	@echo "  VERSION                 Version string (default: git describe)"
	@echo ""
