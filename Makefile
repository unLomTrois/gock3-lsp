# Makefile for gock3-lsp

# Variables
BINARY_NAME=gock3-lsp
CMD_DIR=./cmd/gock3-lsp
BIN_DIR=./bin

# Phony targets to avoid conflicts with files named 'run', 'build', etc.
.PHONY: run build clean test

# Default target
.DEFAULT_GOAL := build

# Build the executable and place it in the bin directory
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)/main.go
	@echo "Build completed. Binary is located at $(BIN_DIR)/$(BINARY_NAME)"

# Run the executable
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
	@echo "  build    - Build the executable"
	@echo "  run      - Build and run the executable"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run tests"
	@echo "  install  - Install the binary to GOPATH/bin"
	@echo "  fmt      - Format the code"
	@echo "  lint     - Lint the code"
