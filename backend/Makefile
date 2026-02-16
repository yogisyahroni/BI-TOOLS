# ============================================================================
# Backend — Makefile for Development & CI
# ============================================================================

.PHONY: test test-coverage test-coverage-html lint build clean

# Default Go test flags
TEST_FLAGS ?= -v -race -count=1

# ──────────────────────────────────────────────────────────────────────
# Testing
# ──────────────────────────────────────────────────────────────────────

## Run all tests
test:
	go test $(TEST_FLAGS) ./...

## Run tests with coverage report (text output)
test-coverage:
	go test $(TEST_FLAGS) -coverprofile=coverage.out -covermode=atomic ./...
	@echo ""
	@echo "============================================"
	@echo "  COVERAGE SUMMARY"
	@echo "============================================"
	go tool cover -func=coverage.out
	@echo "============================================"
	@TOTAL=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}'); \
	echo "Total Coverage: $$TOTAL"

## Run tests with coverage and open HTML report
test-coverage-html: test-coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
ifeq ($(OS),Windows_NT)
	start coverage.html
else
	open coverage.html 2>/dev/null || xdg-open coverage.html 2>/dev/null || true
endif

# ──────────────────────────────────────────────────────────────────────
# Linting
# ──────────────────────────────────────────────────────────────────────

## Run Go linters
lint:
	go vet ./...
	golangci-lint run

## Run Go linters and fix issues
lint-fix:
	golangci-lint run --fix

# ──────────────────────────────────────────────────────────────────────
# Build
# ──────────────────────────────────────────────────────────────────────

## Build the backend binary
build:
	go build -o bin/backend ./main.go

## Clean build artifacts and coverage files
clean:
	rm -rf bin/ coverage.out coverage.html

# ──────────────────────────────────────────────────────────────────────
# Help
# ──────────────────────────────────────────────────────────────────────

## Show available targets
help:
	@echo "Available targets:"
	@echo "  test                - Run all tests"
	@echo "  test-coverage       - Run tests with coverage report"
	@echo "  test-coverage-html  - Run tests and open HTML coverage report"
	@echo "  lint                - Run Go linters"
	@echo "  lint-fix            - Run Go linters with auto-fix"
	@echo "  build               - Build the backend binary"
	@echo "  clean               - Clean build artifacts"
