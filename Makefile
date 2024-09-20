# Makefile for gock3-lsp

# Variables
BINARY_NAME=gock3-lsp
CMD_DIR=./cmd/gock3-lsp
BIN_DIR=./bin

# Supported Platforms
PLATFORMS := linux darwin windows

# Phony targets to avoid conflicts with files named 'run', 'build', etc.
.PHONY: run build clean test install fmt lint help build-linux build-darwin build-windows build-all

# Default target
.DEFAULT_GOAL := build

# Build the executable for the current platform and place it in the bin directory
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BIN_DIR)
	@GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go build -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "Build completed. Binary is located at $(BIN_DIR)/$(BINARY_NAME)"

# Build the executable for Linux
build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BIN_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-linux $(CMD_DIR)/main.go
	@echo "Build completed. Binary is located at $(BIN_DIR)/$(BINARY_NAME)-linux"

# Build the executable for macOS
build-darwin:
	@echo "Building $(BINARY_NAME) for macOS..."
	@mkdir -p $(BIN_DIR)
	@GOOS=darwin GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-darwin $(CMD_DIR)/main.go
	@echo "Build completed. Binary is located at $(BIN_DIR)/$(BINARY_NAME)-darwin"

# Build the executable for Windows
build-windows:
	@echo "Building $(BINARY_NAME) for Windows..."
	@mkdir -p $(BIN_DIR)
	@GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-windows.exe $(CMD_DIR)/main.go
	@echo "Build completed. Binary is located at $(BIN_DIR)/$(BINARY_NAME)-windows.exe"

# Build executables for all supported platforms
build-all: build-linux build-darwin build-windows
	@echo "All platform builds completed."

# Run the executable for the current platform
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BIN_DIR)/$(BINARY_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@echo "Clean completed."

# Run tests
test:
	@echo "Running tests..."
	go test ./...
	@echo "Tests completed."

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to $(GOPATH)/bin..."
	@cp $(BIN_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installation completed."

# Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatting completed."

# Lint the code (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run
	@echo "Linting completed."

# Help target to display available commands
help:
	@echo "Available targets:"
	@echo "  build          - Build the executable for the current platform"
	@echo "  build-linux    - Build the executable for Linux"
	@echo "  build-darwin   - Build the executable for macOS"
	@echo "  build-windows  - Build the executable for Windows"
	@echo "  build-all      - Build executables for all supported platforms"
	@echo "  run            - Build and run the executable for the current platform"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run tests"
	@echo "  install        - Install the binary to GOPATH/bin"
	@echo "  fmt            - Format the code"
	@echo "  lint           - Lint the code"
	@echo "  help           - Show this help message"
