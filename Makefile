.PHONY: help install-tools sqlc migrate-up migrate-down migrate-create migrate-status test build clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-tools: ## Install required development tools (sqlc, dbmate)
	@echo "Installing sqlc..."
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "Installing dbmate..."
	@go install github.com/amacneil/dbmate/v2@latest
	@echo "✓ Tools installed"

sqlc: ## Generate Go code from SQL queries using sqlc
	@echo "Generating SQLC code..."
	@sqlc generate
	@echo "✓ SQLC code generated in internal/sqlc/"

migrate-up: ## Run database migrations
	@dbmate up

migrate-down: ## Rollback last migration
	@dbmate down

migrate-create: ## Create a new migration (usage: make migrate-create NAME=create_users)
	@dbmate new $(NAME)

migrate-status: ## Show migration status
	@dbmate status

test: ## Run all tests
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

test-short: ## Run short tests (no integration tests)
	@go test -short -v ./...

build: ## Build the binary
	@go build -v -o go-annotation ./cmd/go-annotation

clean: ## Clean build artifacts
	@rm -f go-annotation coverage.out coverage.html
	@rm -rf internal/sqlc
	@echo "✓ Cleaned"

gen: sqlc ## Generate all code (alias for sqlc)

.DEFAULT_GOAL := help
