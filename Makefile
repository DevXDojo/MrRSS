# Makefile for MrRSS (Go backend + native macOS SwiftUI client)
.PHONY: help dev build build-app run test test-client test-backend test-coverage \
        lint lint-client lint-backend format format-backend install-deps update-deps \
        check setup clean swagger swagger-validate static-check pre-commit release-check love

SWIFT_PACKAGE := frontend

help: ## Show this help message
	@echo "MrRSS Development Makefile (macOS client)"
	@echo ""
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Development
dev: ## Start the backend and the macOS client together
	./$(SWIFT_PACKAGE)/run.sh

serve: ## Start only the backend on 127.0.0.1:1234
	MRRSS_DEBUG=1 go run . -host 127.0.0.1 -port 1234

# Building
build: build-backend build-client ## Build the backend and the macOS client

build-backend: ## Build the backend binary
	go build -v -o bin/mrrss-server .

build-client: ## Build the macOS client executable
	swift build --package-path $(SWIFT_PACKAGE)

build-app: ## Build the signed .app bundle and DMG (VERSION=x.y.z)
	./$(SWIFT_PACKAGE)/build-app.sh $(or $(VERSION),dev)

run: build-backend ## Run the backend binary
	./bin/mrrss-server -host 127.0.0.1 -port 1234

# Testing
test: test-backend test-client ## Run all tests

test-backend: ## Run backend tests
	go test -v -timeout=5m -cover ./internal/...

test-client: ## Run macOS client tests
	swift test --package-path $(SWIFT_PACKAGE)

test-coverage: ## Run backend tests with coverage
	go test -v -timeout=5m -coverprofile=coverage.out -covermode=atomic ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality
lint: lint-backend lint-client ## Run all linters

lint-backend: ## Run backend linter
	go vet ./...
	find . -name '*.go' -not -path './docs/SERVER_MODE/*' -exec gofmt -d {} + | tee /dev/stderr | test -z "$$(cat)"
	find . -name '*.go' -not -path './docs/SERVER_MODE/*' -exec goimports -d {} + | tee /dev/stderr | test -z "$$(cat)"

lint-client: ## Check macOS client formatting
	@if command -v swift-format >/dev/null 2>&1; then \
		swift-format lint --recursive $(SWIFT_PACKAGE)/Sources $(SWIFT_PACKAGE)/Tests; \
	else \
		echo "swift-format not installed, skipping"; \
	fi

format: format-backend ## Format all code

format-backend: ## Format backend code
	find . -name '*.go' -not -path './docs/SERVER_MODE/*' -exec gofmt -w {} +
	find . -name '*.go' -not -path './docs/SERVER_MODE/*' -exec goimports -w {} +

static-check: ## Run staticcheck for Go code analysis
	staticcheck ./...

# Dependencies
install-deps: ## Install backend dependencies
	go mod download

update-deps: ## Update backend dependencies
	go get -u ./...
	go mod tidy

setup: install-deps ## Initial project setup
	pre-commit install

# Clean
clean: ## Clean build artifacts
	rm -rf bin coverage.out coverage.html $(SWIFT_PACKAGE)/.build $(SWIFT_PACKAGE)/dist
	@echo "Cleaned build artifacts"

check: lint test build ## Run full check (lint, test, build)
	./scripts/check.sh

pre-commit: ## Run pre-commit hooks on all files
	pre-commit run --all-files

release-check: check ## Run all checks before release
	./scripts/pre-release.sh

# API documentation
swagger: ## Generate Swagger API documentation (JSON only)
	swag init -g main.go --parseDependency --parseInternal -o docs/SERVER_MODE
	$(RM) docs/SERVER_MODE/docs.go docs/SERVER_MODE/swagger.yaml
	@echo "Swagger JSON documentation generated: docs/SERVER_MODE/swagger.json"

swagger-validate: ## Validate Swagger annotations
	swag init -g main.go --parseDependency --parseInternal -o docs/SERVER_MODE

love: ## Show some love
	@echo "MrRSS loves you too!"
