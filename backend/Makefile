.PHONY: help run dev build clean test docker-up docker-down migrate migrate-reset seed seed-reset

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

test: ## Run tests
	go test -v ./...

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
