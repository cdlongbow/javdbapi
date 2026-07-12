.PHONY: fmt lint fix test race integration check

fmt:
	@echo "Formatting Go files..."
	@go fmt ./...
	@echo "Formatting complete"

lint:
	@echo "Running linters..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Install with: brew install golangci-lint"; exit 1; }
	@golangci-lint run ./...
	@echo "Linting complete"

fix:
	@echo "Applying automatic fixes..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Install with: brew install golangci-lint"; exit 1; }
	@golangci-lint run ./... --fix
	@go fmt ./...
	@echo "Automatic fixes complete"

test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "Tests complete"

race:
	@go test -race ./...

integration:
	@JAVDB_INTEGRATION=1 go test -tags=integration ./internal/scrape/...

check: lint test race
