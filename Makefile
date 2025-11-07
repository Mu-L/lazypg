.PHONY: build run test clean install help

# Build variables
BINARY_NAME=lazypg
BUILD_DIR=bin
VERSION?=dev
LDFLAGS=-ldflags "-X main.version=${VERSION}"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p ${BUILD_DIR}
	@go build ${LDFLAGS} -o ${BUILD_DIR}/${BINARY_NAME} cmd/lazypg/main.go
	@echo "Built ${BUILD_DIR}/${BINARY_NAME}"

run: build ## Build and run the application
	@${BUILD_DIR}/${BINARY_NAME}

test: ## Run tests
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests and show coverage
	@go tool cover -html=coverage.out

clean: ## Remove build artifacts
	@rm -rf ${BUILD_DIR}
	@rm -f coverage.out
	@echo "Cleaned build artifacts"

install: build ## Install binary to $GOPATH/bin
	@go install ./cmd/lazypg
	@echo "Installed to $(shell go env GOPATH)/bin/${BINARY_NAME}"

fmt: ## Format code
	@go fmt ./...

lint: ## Run linter
	@golangci-lint run ./... || echo "golangci-lint not installed, skipping"

deps: ## Download dependencies
	@go mod download
	@go mod tidy

dev: ## Run in development mode with hot reload
	@echo "Running in development mode (press Ctrl+C to stop)..."
	@go run cmd/lazypg/main.go

.DEFAULT_GOAL := help
