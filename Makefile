# Makefile for MrRSS (Wails v3 + Task)
.PHONY: help dev build package run test test-frontend test-backend lint lint-frontend format format-backend install-deps update-deps check setup clean love swagger swagger-validate swagger-serve static-check

# Detect OS
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    SHELL := pwsh.exe
    .SHELLFLAGS := -Command
    TASK := task.exe
else
    DETECTED_OS := $(shell uname -s)
    SHELL := /bin/bash
    TASK := task
endif

# Default target
help: ## Show this help message
	@echo "MrRSS Development Makefile ($(DETECTED_OS))"
	@echo ""
	@echo "Wails v3 Build System - Using Task Runner"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "💡 Tip: Use 'task --list' to see all available tasks"

# Development (Wails v3 + Task)
dev: ## Start development server with hot reload
	$(TASK) dev

# Building (Wails v3 + Task)
build: ## Build application for current platform
	$(TASK) build

package: ## Package application with installer
	$(TASK) package

run: ## Run the built application
	$(TASK) run

build-frontend: ## Build frontend only
	$(TASK) common:build:frontend

build-backend: ## Build backend only
	go build -v -o build/bin/ ./...

# Testing
test: test-frontend test-backend ## Run all tests

test-frontend: ## Run frontend tests
	cd frontend && npm test

test-frontend-e2e: ## Run frontend E2E tests with Cypress
	cd frontend && npm run test:e2e

test-backend: ## Run backend tests
	go test -v -timeout=5m -cover ./internal/...

test-coverage: ## Run backend tests with coverage
	go test -v -timeout=5m -coverprofile=coverage.out -covermode=atomic ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-all: test-frontend test-frontend-e2e test-backend ## Run all tests including E2E

# Code Quality
lint: lint-frontend lint-backend ## Run all linters

lint-frontend: ## Run frontend linter
	cd frontend && npm run lint

lint-backend: ## Run backend linter
	go vet ./...
ifeq ($(DETECTED_OS),Windows)
	powershell -Command '$$files = Get-ChildItem -Recurse -Include *.go | Where-Object { $$_ -notmatch "docs[\\/]SERVER_MODE" }; $$result = gofmt -d $$files ; if ($$result) { Write-Host $$result -ForegroundColor Red; exit 1 }'
	powershell -Command '$$files = Get-ChildItem -Recurse -Include *.go | Where-Object { $$_ -notmatch "docs[\\/]SERVER_MODE" }; $$importsResult = goimports -d $$files ; if ($$importsResult) { Write-Host $$importsResult -ForegroundColor Red; exit 1 }'
else
	find . -name '*.go' -not -path './docs/SERVER_MODE/*' -exec gofmt -d {} + | tee /dev/stderr | test -z "$$(cat)"
	find . -name '*.go' -not -path './docs/SERVER_MODE/*' -exec goimports -d {} + | tee /dev/stderr | test -z "$$(cat)"
endif

format: format-frontend format-backend ## Format all code

format-frontend: ## Format frontend code
	cd frontend && npm run format

format-backend: ## Format backend code
ifeq ($(DETECTED_OS),Windows)
	powershell -Command '$$files = Get-ChildItem -Recurse -Include *.go | Where-Object { $$_ -notmatch "docs[\\/]SERVER_MODE" }; gofmt -w $$files; goimports -w $$files'
else
	find . -name '*.go' -not -path './docs/SERVER_MODE/*' -exec gofmt -w {} +
	find . -name '*.go' -not -path './docs/SERVER_MODE/*' -exec goimports -w {} +
endif

# Dependencies
install-deps: install-frontend-deps install-backend-deps ## Install all dependencies

install-frontend-deps: ## Install frontend dependencies
	cd frontend && npm install

install-backend-deps: ## Install backend dependencies
	go mod download

update-deps: update-frontend-deps update-backend-deps ## Update all dependencies

update-frontend-deps: ## Update frontend dependencies
	cd frontend && npm update

update-backend-deps: ## Update backend dependencies
	go get -u ./...
	go mod tidy

# Setup
setup: install-deps ## Initial project setup
	pre-commit install

# Task runner commands
task-list: ## List all available tasks
	$(TASK) --list

task-summary: ## Show task summary
	$(TASK) --summary build dev package

icons: ## Generate platform icons
	$(TASK) common:generate:icons

bindings: ## Generate TypeScript bindings
	$(TASK) common:generate:bindings

setup-docker: ## Setup Docker for cross-compilation
	$(TASK) common:setup:docker

# Clean
clean: ## Clean build artifacts
ifeq ($(DETECTED_OS),Windows)
	-Remove-Item -Recurse -Force build/bin,frontend/dist,coverage.out,coverage.html,*.syso 2>$$null
else
	rm -rf build/bin frontend/dist coverage.out coverage.html *.syso
endif
	@echo "✅ Cleaned build artifacts"

# Development helpers
check: lint test build ## Run full check (lint, test, build)
ifeq ($(DETECTED_OS),Windows)
	powershell -File scripts/check.ps1
else
	./scripts/check.sh
endif

pre-commit: ## Run pre-commit hooks on all files
	pre-commit run --all-files

release-check: check ## Run all checks before release
ifeq ($(DETECTED_OS),Windows)
	powershell -File scripts/pre-release.ps1
else
	./scripts/pre-release.sh
endif

love: ## Show some love
	@echo "❤️ MrRSS loves you too! ❤️"

# Platform-specific builds
build-windows: ## Build for Windows
	$(TASK) windows:build

build-linux: ## Build for Linux
	$(TASK) linux:build

build-darwin: ## Build for macOS
	$(TASK) darwin:build

# API Documentation
swagger: ## Generate Swagger API documentation (JSON only)
	swag init -g main-core.go --parseDependency --parseInternal -o docs/SERVER_MODE
ifeq ($(DETECTED_OS),Windows)
	powershell -Command "Remove-Item -ErrorAction SilentlyContinue docs/SERVER_MODE/docs.go, docs/SERVER_MODE/swagger.yaml"
else
	$(RM) docs/SERVER_MODE/docs.go docs/SERVER_MODE/swagger.yaml
endif
	@echo "✅ Swagger JSON documentation generated: docs/SERVER_MODE/swagger.json"

swagger-validate: ## Validate Swagger annotations
	@echo "🔍 Validating Swagger annotations..."
	swag init -g main-core.go --parseDependency --parseInternal --parseInternal -o docs/SERVER_MODE

swagger-serve: ## Generate docs and serve Swagger UI
	swag init -g main-core.go --parseDependency --parseInternal -o docs/SERVER_MODE
	@echo "🌐 Starting development server..."
	$(TASK) dev

static-check: ## Run staticcheck for Go code analysis
	staticcheck ./...
