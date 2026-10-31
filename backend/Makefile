.PHONY: help run dev build clean test test-integration test-db test-integration-isolated docker-up docker-down migrate migrate-reset seed seed-reset

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

run: ## Run the API server
	go run cmd/api/main.go

dev: ## Run with air for hot reload (requires air)
	air

build: ## Build the API binary
	@echo "Building API server..."
	@go build -o bin/api cmd/api/main.go
	@echo "Build completed!"

clean: ## Remove built binaries
	@rm -rf bin/
	@echo "Cleaned!"

test: ## Run unit tests (no database needed)
	go test -v ./...

TEST_DB_CONTAINER ?= the-zins-db
TEST_DB_NAME ?= the_zins_db_test

test-integration: ## Integration tests on the dev DB in .env (each test rolls back; schema must be migrated)
	go test -tags integration -count=1 -v ./test/integration/...

test-db: ## Create the separate test database inside the Docker container (idempotent)
	docker-compose -f ../docker-compose.yml up -d db
	@until docker exec $(TEST_DB_CONTAINER) pg_isready -U postgres -q; do sleep 1; done
	@docker exec $(TEST_DB_CONTAINER) psql -U postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$(TEST_DB_NAME)'" | grep -q 1 \
		|| docker exec $(TEST_DB_CONTAINER) createdb -U postgres $(TEST_DB_NAME)

test-integration-isolated: test-db ## Integration tests on $(TEST_DB_NAME) (auto-migrated, never touches dev data)
	INTEGRATION_DB_NAME=$(TEST_DB_NAME) go test -tags integration -count=1 -v ./test/integration/...

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

migrate: ## Run database migrations
	@echo "Running migrations..."
	@go run cmd/api/main.go

migrate-reset: ## Reset and run migrations
	@echo "Resetting database..."
	@echo "This will drop all tables and re-run migrations"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "Not implemented yet - add reset logic to your app"; \
	fi

seed: ## Seed database with sample data
	@echo "Seeding database..."
	go run cmd/seed/main.go

seed-reset: ## Reset database and seed with sample data
	@echo "Resetting and seeding database..."
	go run cmd/seed/main.go -reset
