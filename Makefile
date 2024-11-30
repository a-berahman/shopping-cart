DOCKER_COMPOSE := docker-compose.yml
CONFIG_FILE := config.yaml

MIGRATIONS_DIR=migrations
DATABASE_URL ?= postgres://postgres:postgres@shopping_cart_postgres:5432/shopping_cart?sslmode=disable

.PHONY: run
run:
	@echo "Starting application..."
	@if [ ! -f $(CONFIG_FILE) ] && [ -f config.yaml.example ]; then \
		echo "Creating config file from example..."; \
		cp config.yaml.example $(CONFIG_FILE); \
	fi
	@docker-compose -f $(DOCKER_COMPOSE) up -d
	@echo "Application is running. Use 'make logs' to view logs"
	docker-compose logs -f app

.PHONY: stop
stop:
	@echo "Stopping application..."
	@docker-compose -f $(DOCKER_COMPOSE) down --volumes
	@echo "Appli

.PHONY: logs
logs:
	@docker-compose -f $(DOCKER_COMPOSE) logs -f

.PHONY: migrate-up
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down

.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

.PHONY: migrate-force
migrate-force:
	@read -p "Enter version number: " version; \
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $$version

.PHONY: migrate-version
migrate-version:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

.PHONY: sqlc
sqlc:
	sqlc generate

# Development commands
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@docker-compose -f $(DOCKER_COMPOSE) down --volumes --remove-orphans
	@docker system prune -f
	@echo "Cleanup complete"

.PHONY: coverage
coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@rm coverage.out

