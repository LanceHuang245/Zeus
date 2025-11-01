# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary names
BINARY_NAME=zeus
BINARY_UNIX=$(BINARY_NAME)_unix

# Build flags
LDFLAGS=-ldflags "-X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')"

# Version info
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')
BUILD_TIME=$(shell date +%Y-%m-%dT%H:%M:%S%z)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')

.PHONY: help build build-linux build-windows clean tidy test run deps

# Default target
all: help

# Show this help message
help: ## Display this help screen
	@echo 'Available commands:'
	@echo 'Usage: make [target]'
	@echo ''
	@echo '  build              Build the binary for current platform'
	@echo '  build-linux        Build the binary for Linux'
	@echo '  build-windows      Build the binary for Windows'
	@echo '  build-all          Build the binary for all platforms'
	@echo '  clean              Clean build artifacts'
	@echo '  tidy               Tidy Go module dependencies'
	@echo '  test               Run tests'
	@echo '  run                Run the application locally'
	@echo '  deps               Download dependencies'
	@echo '  fmt                Format Go code'
	@echo '  help               Show this help message'
	@echo ''
	@echo 'Example usage:'
	@echo '  make build-linux'
	@echo '  make build-windows'
	@echo '  make tidy'

# Build for current platform
build: ## Build the binary for current platform
	@echo "Building $(BINARY_NAME) for current platform..."
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) -v
	@echo "Build completed: bin/$(BINARY_NAME)"

# Build for Linux
build-linux: ## Build the binary for Linux
	@echo "Building $(BINARY_NAME) for Linux..."
	@echo "Version: $(VERSION)"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 -v
	@echo "Linux build completed: bin/$(BINARY_NAME)-linux-amd64"

# Build for Windows
build-windows: ## Build the binary for Windows
	@echo "Building $(BINARY_NAME) for Windows..."
	@echo "Version: $(VERSION)"
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe -v
	@echo "Windows build completed: bin/$(BINARY_NAME)-windows-amd64.exe"

# Build for all platforms
build-all: build-linux build-windows ## Build the binary for all platforms
	@echo "All builds completed"

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)-*
	rm -f bin/$(BINARY_NAME)*
	@echo "Clean completed"

# Tidy module dependencies
tidy: ## Tidy Go module dependencies
	@echo "Tidying Go modules..."
	$(GOMOD) tidy
	$(GOMOD) verify
	@echo "Go modules tidied"

# Download dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded"

# Run the application
run: ## Run the application locally
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run main.go

# Format Go code
fmt: ## Format Go code
	@echo "Formatting Go code..."
	$(GOCMD) fmt ./...
	@echo "Code formatted"
