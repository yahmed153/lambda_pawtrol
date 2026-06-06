# ====================================================================================
# Variables
# ====================================================================================
BINARY_NAME=bootstrap
BUILD_DIR=bin
MAIN_PACKAGE_PATH=./main.go
DEPLOYMENT_PACKAGE=deployment.zip

# ====================================================================================
# Default Target
# ====================================================================================
.DEFAULT_GOAL := help

# ====================================================================================
# Development Targets
# ====================================================================================

## build: Alias for build-linux-arm64 target
.PHONY: build
build: build-linux-arm64

## run: Build and execute the binary locally
.PHONY: run
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

## clean: Remove build artifacts and temporary directories
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DEPLOYMENT_PACKAGE)
	@echo "✨ Clean complete"

## fmt: Format all Go source files using go fmt
.PHONY: fmt
fmt:
	@echo "🎨 Formatting source files..."
	go fmt ./...

## test: Run unit tests with the race detector enabled
.PHONY: test
test:
	@echo "🧪 Running unit tests..."
	go test -race -v ./...

# ====================================================================================
# Cross-Compilation Targets
# ====================================================================================

## build-linux-arm64: Compile the binary for Linux (arm64)
.PHONY: build-linux-arm64
build-linux-arm64:
	@echo "🐧 Building for Linux (arm64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE_PATH)

## build-linux-amd64: Compile the binary for Linux (amd64)
.PHONY: build-linux-amd64
build-linux-amd64:
	@echo "🐧 Building for Linux (arm64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE_PATH)

## package: Package binary to be uploaded to AWS Lambda
.PHONY: package
package: build-linux-arm64
	@echo "🐧 Packaging binary for Linux (arm64)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) .
	@zip deployment.zip $(BINARY_NAME)
	@rm -f $(BINARY_NAME)

# ====================================================================================
# Help
# ====================================================================================

## help: Show this help message with available targets
.PHONY: help
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' |  sed -e 's/^/ /'
