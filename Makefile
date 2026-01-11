.PHONY: help proto build test lint fmt clean install release

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

proto: ## Generate Go code from protobuf definitions
	@echo "Generating protobuf files..."
	@protoc --go_out=. --go_opt=module=github.com/conallob/jira-beads-sync --proto_path=proto proto/jira.proto proto/beads.proto
	@echo "Protobuf generation complete"

build: proto ## Build the binary
	@echo "Building jira-beads-sync..."
	@go build -o jira-beads-sync ./cmd/jira-beads-sync
	@echo "Build complete: ./jira-beads-sync"

test: proto ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

coverage: test ## Show test coverage
	@go tool cover -html=coverage.out

lint: proto ## Run linter
	@echo "Running linter..."
	@golangci-lint run
	@echo "Linting complete"

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Formatting complete"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f jira-beads-sync
	@rm -f coverage.out
	@rm -rf dist/
	@rm -rf .beads/
	@echo "Clean complete"

install: build ## Install binary to $GOPATH/bin
	@echo "Installing to $(GOPATH)/bin..."
	@cp jira-beads-sync $(GOPATH)/bin/
	@echo "Installation complete"

release-dry-run: proto ## Test release process without publishing
	@echo "Running GoReleaser in dry-run mode..."
	@goreleaser release --snapshot --clean
	@echo "Dry-run complete. Check dist/ directory"

release-snapshot: proto ## Create a snapshot release (no tag required)
	@echo "Creating snapshot release..."
	@goreleaser release --snapshot --clean
	@echo "Snapshot release complete"

deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

verify: proto fmt lint test ## Run all verification steps
	@echo "All verification steps passed!"
