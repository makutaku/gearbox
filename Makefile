# Gearbox Build System
.PHONY: all build clean test install dev-setup cli legacy-build

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION = $(shell go version | cut -d' ' -f3)

# Build flags
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.goVersion=$(GO_VERSION)"

# Default target
all: build

# Build all components
build: cli tools

# Build the Go CLI
cli:
	@echo "Building gearbox CLI..."
	cd cmd/gearbox && go build $(LDFLAGS) -o ../../build/gearbox

# Build Go tools from unified module
tools: build/orchestrator build/script-generator build/config-manager

build/orchestrator: cmd/orchestrator/main.go pkg/orchestrator/main.go
	@echo "Building orchestrator..."
	@mkdir -p build
	@go build -o build/orchestrator ./cmd/orchestrator

build/script-generator: cmd/script-generator/main.go pkg/generator/main.go
	@echo "Building script-generator..."
	@mkdir -p build
	@go build -o build/script-generator ./cmd/script-generator

build/config-manager: cmd/config-manager/main.go pkg/config/main.go
	@echo "Building config-manager..."
	@mkdir -p build
	@go build -o build/config-manager ./cmd/config-manager

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	go mod tidy
	go mod download

# Development setup
dev-setup: deps build
	@echo "Setting up development environment..."

# Testing
test:
	@echo "Running Go tests..."
	go test ./...
	
	@echo "Running shell tests..."
	@if [ -f test.sh ]; then ./test.sh; fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf build/

# Install system-wide (requires sudo)
install: build
	@echo "Installing gearbox system-wide..."
	sudo cp build/gearbox /usr/local/bin/gearbox
	@echo "Gearbox CLI installed to /usr/local/bin/gearbox"

# Development shortcuts
dev: dev-setup
	@echo "Development environment ready!"
	@echo "Test the CLI with: ./gearbox --help"

# Quick development test
dev-test: build
	@echo "Testing CLI..."
	@./build/gearbox --help

# Show build information
info:
	@echo "Build Information:"
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Go Version: $(GO_VERSION)"

# Help
help:
	@echo "Gearbox Build System"
	@echo ""
	@echo "Targets:"
	@echo "  all          Build everything (default)"
	@echo "  build        Build CLI and tools"
	@echo "  cli          Build only the Go CLI"
	@echo "  tools        Build only the Go tools"
	@echo "  deps         Install Go dependencies"
	@echo "  dev-setup    Setup development environment"
	@echo "  test         Run all tests"
	@echo "  clean        Clean build artifacts"
	@echo "  install      Install system-wide (requires sudo)"
	@echo "  dev          Quick development setup and test"
	@echo "  info         Show build information"
	@echo "  help         Show this help"