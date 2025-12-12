# Makefile for MrRSS (Cross-platform)
.PHONY: help build build-frontend build-backend test test-frontend test-backend lint lint-frontend format format-frontend clean dev setup install-deps update-deps check pre-commit release-check

# Detect OS
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    SHELL := pwsh.exe
    .SHELLFLAGS := -Command
    EXTRA_BUILD_FLAGS :=
else
    DETECTED_OS := $(shell uname -s)
    SHELL := /bin/bash
    ifeq ($(DETECTED_OS),Darwin)
        # macOS uses -tags nosystray to exclude fyne.io/systray (conflicts with Wails AppDelegate)
        # macOS uses DarwinKit for system tray which is selected by build constraints
        EXTRA_BUILD_FLAGS := -tags nosystray
    else
        EXTRA_BUILD_FLAGS :=
    endif
endif

# Default target
help: ## Show this help message
	@echo "MrRSS Development Makefile ($(DETECTED_OS))"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Development
dev: ## Start development server (auto-detects platform and applies correct tags)
ifeq ($(DETECTED_OS),Darwin)
	@echo "🚀 Starting Wails dev for macOS with -tags nosystray..."
	wails dev -tags nosystray
else
	@echo "🚀 Starting Wails dev..."
	wails dev
endif

# Building
build: build-frontend build-backend ## Build both frontend and backend
	wails build -skipbindings $(EXTRA_BUILD_FLAGS)

build-frontend: ## Build frontend only
	cd frontend && npm run build

build-backend: ## Build backend only (verify compilation)
	go build -v ./...

# Testing
test: test-frontend test-backend ## Run all tests

test-frontend: ## Run frontend tests
	cd frontend && npm test

test-backend: ## Run backend tests
	go test -v -timeout=5m -cover ./internal/...

test-coverage: ## Run backend tests with coverage
	go test -v -timeout=5m -coverprofile=coverage.out -covermode=atomic ./internal/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code Quality
lint: lint-frontend lint-backend ## Run all linters

lint-frontend: ## Run frontend linter
	cd frontend && npm run lint

lint-backend: ## Run backend linter
	go vet ./...
ifeq ($(DETECTED_OS),Windows)
	powershell -Command '$$result = gofmt -d . ; if ($$result) { Write-Host $$result -ForegroundColor Red; exit 1 }'
	powershell -Command '$$importsResult = goimports -d . ; if ($$importsResult) { Write-Host $$importsResult -ForegroundColor Red; exit 1 }'
else
	gofmt -d . | tee /dev/stderr | test -z "$$(cat)"
	goimports -d . | tee /dev/stderr | test -z "$$(cat)"
endif

format: format-frontend format-backend ## Format all code

format-frontend: ## Format frontend code
	cd frontend && npm run format

format-backend: ## Format backend code
	gofmt -w .
	goimports -w .

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

# Cleanup
clean: ## Clean build artifacts
ifeq ($(DETECTED_OS),Windows)
	powershell -Command "Remove-Item -Path 'build\bin\*' -Recurse -Force -ErrorAction SilentlyContinue"
	powershell -Command "Remove-Item -Path 'frontend\dist' -Recurse -Force -ErrorAction SilentlyContinue"
	powershell -Command "Remove-Item -Path 'frontend\node_modules\.vite' -Recurse -Force -ErrorAction SilentlyContinue"
	powershell -Command "Remove-Item -Path 'coverage.out', 'coverage.html' -ErrorAction SilentlyContinue"
else
	rm -rf build/bin/*
	rm -rf frontend/dist
	rm -rf frontend/node_modules/.vite
	rm -f coverage.out coverage.html
endif

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

# Docker (if needed in future)
# docker-build:
# 	docker build -t mrrss .

# docker-run:
# 	docker run -p 3000:3000 mrrss
