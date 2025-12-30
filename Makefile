.PHONY: help build test lint clean run docker-build docker-run install coverage docs

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

VERSION := $(shell cat VERSION 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-w -s -X main.version=$(VERSION)"

docs: ## Generate Swagger documentation
	@echo "Generating Swagger docs..."
	@command -v swag >/dev/null 2>&1 || go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/messagarr/main.go -o docs --parseDependency --parseInternal

build: docs ## Build the application (regenerates docs first)
	@echo "Building Messagarr v$(VERSION)..."
	go build $(LDFLAGS) -o messagarr ./cmd/messagarr

install: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

test: ## Run tests
	@echo "Running tests..."
	go test -v -race ./...

coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f messagarr
	rm -f coverage.out coverage.html
	go clean

run: ## Run the application
	@echo "Running Messagarr..."
	go run ./cmd/messagarr

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t messagarr:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker-compose up -d

docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f messagarr

format: ## Format code
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

check: lint vet test ## Run all checks (lint, vet, test)

release: clean check build ## Prepare for release

.DEFAULT_GOAL := help
