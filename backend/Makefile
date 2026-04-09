.PHONY: build test clean migrate-up migrate-down migrate-create run dev lint fmt help

# Variables
APP_NAME=api
BUILD_DIR=bin
MIGRATE_CMD=migrate
DB_URL?=$(shell grep DB_URL ../.env 2>/dev/null | cut -d '=' -f2)

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo "Available commands:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'

## build: Build the Go application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/api
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

## build-worker: Build the notification worker
build-worker:
	@echo "Building notification worker..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/worker ./cmd/worker
	@echo "Build complete: $(BUILD_DIR)/worker"

## test: Run all tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

## test-short: Run short tests (skip integration)
test-short:
	@echo "Running short tests..."
	SKIP_INTEGRATION_TESTS=true go test -v -race -short ./...

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration/...

## test-contract: Run API contract tests
test-contract:
	@echo "Running contract tests..."
	go test -v ./tests/contract/...

## test-benchmark: Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./tests/benchmark/...

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | tail -1
	@echo "Coverage report: coverage.html"

## security-scan: Run security scans
test-security:
	@echo "Running security scans..."
	@bash ../scripts/security-scan.sh

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

## migrate-up: Run all pending migrations
migrate-up:
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL not set. Run: make migrate-up DB_URL=postgres://..."; \
		exit 1; \
	fi
	@echo "Running migrations up..."
	$(MIGRATE_CMD) -path migrations -database "$(DB_URL)" up

## migrate-down: Rollback one migration
migrate-down:
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL not set. Run: make migrate-down DB_URL=postgres://..."; \
		exit 1; \
	fi
	@echo "Running migration down..."
	$(MIGRATE_CMD) -path migrations -database "$(DB_URL)" down 1

## migrate-create: Create a new migration file (usage: make migrate-create name=create_users_table)
migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: Migration name required. Run: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)..."
	$(MIGRATE_CMD) create -ext sql -dir migrations -seq $(name)

## migrate-version: Show current migration version
migrate-version:
	@if [ -z "$(DB_URL)" ]; then \
		echo "Error: DB_URL not set"; \
		exit 1; \
	fi
	$(MIGRATE_CMD) -path migrations -database "$(DB_URL)" version

## migrate-force: Force migration version (usage: make migrate-force version=1)
migrate-force:
	@if [ -z "$(DB_URL)" ] || [ -z "$(version)" ]; then \
		echo "Error: DB_URL and version required. Run: make migrate-force version=1"; \
		exit 1; \
	fi
	$(MIGRATE_CMD) -path migrations -database "$(DB_URL)" force $(version)

## run: Run the application (requires .env)
run: build
	@echo "Starting application..."
	@../.env 2>/dev/null && export $$(cat ../.env | xargs) && ./$(BUILD_DIR)/$(APP_NAME)

## dev: Run the application with hot reload (requires air)
dev:
	@which air > /dev/null 2>&1 || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

## lint: Run golangci-lint
lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Formatting complete"

## deps: Download and verify dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify
	@echo "Dependencies ready"

## tidy: Tidy and vendor dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "Dependencies tidied"
