# Chisel Makefile

.PHONY: build test test-integration clean fmt lint install dev docs

# Build variables
BINARY_NAME=chisel
BUILD_DIR=bin
DOCS_DIR=docs
VERSION?=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=${VERSION}"

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Default target
all: test docs build

# Generate documentation
docs:
	@echo "Generating documentation..."
	@mkdir -p $(DOCS_DIR)
	@$(GOCMD) run scripts/generate-docs.go $(DOCS_DIR)/
	@echo "Documentation generated in $(DOCS_DIR)/"

# Build the binary
build: docs
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/chisel

# Development build (faster, no docs)
dev-build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/chisel

# Run tests
test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

# Run integration tests
test-integration:
	$(GOTEST) -v -tags=integration ./tests/integration/...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DOCS_DIR)
	rm -f coverage.out

# Format code
fmt:
	$(GOCMD) fmt ./...

# Lint code
lint:
	golangci-lint run

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install binary to GOPATH/bin
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Development mode - watch for changes and rebuild
dev:
	@echo "Starting development mode..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# Generate code coverage report
coverage: test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	$(GOTEST) -bench=. -benchmem ./...

# Security scan
security:
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...

# Check for vulnerabilities
vuln:
	@which govulncheck > /dev/null || (echo "Installing govulncheck..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...

# Release build for multiple platforms
release: docs
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/chisel
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/chisel
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/chisel
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/chisel
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/chisel

# Docker build
docker:
	docker build -t chisel:$(VERSION) .

# Run examples
example-simple: dev-build
	./$(BUILD_DIR)/$(BINARY_NAME) plan --module examples/simple-file/module.yaml

example-user: dev-build
	./$(BUILD_DIR)/$(BINARY_NAME) plan --module examples/user-management/module.yaml

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary with documentation"
	@echo "  dev-build      - Fast development build (no docs)"
	@echo "  docs           - Generate documentation only"
	@echo "  test           - Run unit tests"
	@echo "  test-integration - Run integration tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  deps           - Install dependencies"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  dev            - Development mode with hot reload"
	@echo "  coverage       - Generate coverage report"
	@echo "  bench          - Run benchmarks"
	@echo "  security       - Run security scan"
	@echo "  vuln           - Check for vulnerabilities"
	@echo "  release        - Build for multiple platforms"
	@echo "  docker         - Build Docker image"
	@echo "  example-simple - Run simple file example"
	@echo "  example-user   - Run user management example"
	@echo "  help           - Show this help"
