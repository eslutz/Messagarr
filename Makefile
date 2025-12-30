.PHONY: help build test lint clean run docker-build docker-run install coverage

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building Messagarr..."
	go build -o messagarr ./cmd/messagarr

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
