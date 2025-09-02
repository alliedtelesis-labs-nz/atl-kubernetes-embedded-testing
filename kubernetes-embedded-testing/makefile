# Build variables
BINARY_NAME=ket
BINARY_UNIX=$(BINARY_NAME)_unix
DOCKER_IMAGE=ket:latest
GO_FILES=$(shell find . -name "*.go" -type f)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_DIR=bin

# Default target
.DEFAULT_GOAL := build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -buildvcs=false -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/testrunner
	@echo "Build complete: $(BINARY_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -buildvcs=false -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/testrunner
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -buildvcs=false -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/testrunner
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -buildvcs=false -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/testrunner
	@echo "Multi-platform build complete"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...
	@echo "Tests complete"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOGET) -v -t -d ./...
	@echo "Dependencies installed"

# Update go modules
.PHONY: mod-update
mod-update:
	@echo "Updating go modules..."
	$(GOMOD) tidy
	$(GOMOD) download
	@echo "Modules updated"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "Code formatted"

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	$(GOCMD) vet ./...
	@echo "Linting complete"

# Build Docker images
.PHONY: docker-build
docker-build:
	@echo "Building Docker images..."
	docker build --target base -t kubernetes-embedded-test-runner-base:latest .
	docker build --target node-runner -t kubernetes-embedded-test-runner-node:latest .
	docker build --target golang-runner -t kubernetes-embedded-test-runner-golang:latest .
	docker build -t $(DOCKER_IMAGE) .
	@echo "All Docker images built:"
	@echo "  - kubernetes-embedded-test-runner-base:latest"
	@echo "  - kubernetes-embedded-test-runner-node:latest"
	@echo "  - kubernetes-embedded-test-runner-golang:latest"
	@echo "  - $(DOCKER_IMAGE)"

# Build specific Docker image
.PHONY: docker-build-base
docker-build-base:
	@echo "Building base Docker image..."
	docker build --target base -t kubernetes-embedded-test-runner-base:latest .
	@echo "Base Docker image built: kubernetes-embedded-test-runner-base:latest"

.PHONY: docker-build-node
docker-build-node:
	@echo "Building Node.js Docker image..."
	docker build --target node-runner -t kubernetes-embedded-test-runner-node:latest .
	@echo "Node.js Docker image built: kubernetes-embedded-test-runner-node:latest"

.PHONY: docker-build-golang
docker-build-golang:
	@echo "Building Go Docker image..."
	docker build --target golang-runner -t kubernetes-embedded-test-runner-golang:latest .
	@echo "Go Docker image built: kubernetes-embedded-test-runner-golang:latest"

# Run Docker image
.PHONY: docker-run
docker-run:
	@echo "Running Docker image..."
	docker run --rm $(DOCKER_IMAGE) --help

# Build and run locally
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_DIR)/$(BINARY_NAME) --help

# Install binary to system
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BINARY_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete"

# Uninstall binary
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Uninstallation complete"

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-all      - Build for multiple platforms"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  deps           - Install dependencies"
	@echo "  mod-update     - Update go modules"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  docker-build   - Build all Docker images"
	@echo "  docker-build-base  - Build base Docker image"
	@echo "  docker-build-node   - Build Node.js Docker image"
	@echo "  docker-build-golang - Build Go Docker image"
	@echo "  docker-run     - Run Docker image"
	@echo "  run            - Build and run locally"
	@echo "  install        - Install binary to system"
	@echo "  uninstall      - Uninstall binary"
	@echo "  help           - Show this help message"
