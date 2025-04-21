# Variables
DEV_CONFIG_FILE := config/config.dev.yml
DEV_COMPOSE_FILE := build/docker-compose/docker-compose-dev.yml

# Extract values from config.dev.yml using yq (you need to install yq first)
DEV_DB_PASSWORD := $(shell yq e '.db.password' $(DEV_CONFIG_FILE))
DEV_DB_NAME := $(shell yq e '.db.dbname' $(DEV_CONFIG_FILE))
DEV_DB_USER := $(shell yq e '.db.user' $(DEV_CONFIG_FILE))
DEV_ARANGO_PASSWORD := $(shell yq e '.arango.password' $(DEV_CONFIG_FILE))
DEV_REDIS_PASSWORD := $(shell yq e '.redis.password' $(DEV_CONFIG_FILE))


# Common environment variables
define set_dev_env_vars
	DB_PASSWORD=$(DEV_DB_PASSWORD) \
	DB_NAME=$(DEV_DB_NAME) \
	DB_USER=$(DEV_DB_USER) \
	ARANGO_PASSWORD=$(DEV_ARANGO_PASSWORD) \
	REDIS_PASSWORD=$(DEV_REDIS_PASSWORD)
endef

# Helper function to wait for services
define wait_for_services
	@echo "Waiting for services to be ready..."
	@sleep 10
endef

.PHONY: dev-build
dev-build:
	docker-compose -f $(DEV_COMPOSE_FILE) build app

# Development commands
.PHONY: dev-up
dev-up:
	@echo "Starting development environment..."
	$(call set_dev_env_vars) docker-compose -f $(DEV_COMPOSE_FILE) up -d app

.PHONY: dev-build-up
dev-build-up: dev-build dev-up


.PHONY: dev-down
dev-down:
	docker-compose -f $(DEV_COMPOSE_FILE) down

.PHONY: dev-logs
dev-logs:
	docker-compose -f $(DEV_COMPOSE_FILE) logs -f

.PHONY: dev-app-logs
dev-app-logs:
	docker-compose -f $(DEV_COMPOSE_FILE) logs -f app

.PHONY: dev-ps
dev-ps:
	docker-compose -f $(DEV_COMPOSE_FILE) ps 


.PHONY: dev-create
dev-create:
	@echo "Starting database creation..."
	$(call set_dev_env_vars) docker-compose -f $(DEV_COMPOSE_FILE) up -d postgres arango
	$(call wait_for_services)
	$(call set_dev_env_vars) docker-compose -f $(DEV_COMPOSE_FILE) up --build db-creator


.PHONY: dev-migrate
dev-migrate:
	@echo "Starting migration..."
	$(call set_dev_env_vars) docker-compose -f $(DEV_COMPOSE_FILE) up -d postgres arango
	$(call wait_for_services)
	$(call set_dev_env_vars) docker-compose -f $(DEV_COMPOSE_FILE) up db-migrator

.PHONY: dev-clear
dev-clear:
	@echo "Starting clearing..."
	@echo "Stopping containers..."
	docker-compose -f $(DEV_COMPOSE_FILE) down
	docker-compose -f $(DEV_COMPOSE_FILE) down --remove-orphans
	@echo "Removing containers and volumes..."
	docker-compose -f $(DEV_COMPOSE_FILE) rm -v -s
	@echo "Removing named volumes..."
	docker-compose -f $(DEV_COMPOSE_FILE) down -v


.PHONY: gen-proto
gen-proto:
	protoc \
	--proto_path=./proto \
	--go_out=./proto \
	--go_opt=paths=source_relative \
	--go-grpc_opt=paths=source_relative \
	--go-grpc_opt=require_unimplemented_servers=false \
	--experimental_allow_proto3_optional \
	--go-grpc_out=./proto \
	./proto/*.proto

.PHONY: help
help:
	@echo ""
	@echo "Available commands:"
	@echo "  dev-build      - Build the application container"
	@echo "  dev-up         - Start the development environment"
	@echo "  dev-build-up   - Build and start the development environment"
	@echo "  dev-down       - Stop the development environment"
	@echo "  dev-logs       - Show logs from all containers"
	@echo "  dev-app-logs   - Show logs from the app container"
	@echo "  dev-ps         - List running containers"
	@echo "  dev-create     - Create and initialize databases"
	@echo "  dev-migrate    - Run database migrations"
	@echo "  dev-clear      - Remove all containers and volumes"
	@echo "  gen-proto      - Generate protobuf code"
	@echo ""

.DEFAULT_GOAL := help
