.PHONY: help
.PHONY: up down docker-build logs restart clean
.PHONY: test test-unit test-integration
.PHONY: build validate deploy-staging deploy-production prepare-schema-init
.PHONY: sam-build sam-validate sam-deploy-staging sam-deploy-production

# =========================
# HELP
# =========================

help:
	@echo "Available commands:"
	@echo ""
	@echo "Local environment:"
	@echo "  make up                  - start local environment"
	@echo "  make down                - stop local environment"
	@echo "  make restart             - restart local environment"
	@echo "  make logs                - show container logs"
	@echo "  make clean               - stop containers and remove volumes"
	@echo ""
	@echo "Testing:"
	@echo "  make test                - run all tests"
	@echo "  make test-unit           - run unit/domain tests"
	@echo "  make test-integration    - run integration tests"
	@echo ""
	@echo "AWS / SAM:"
	@echo "  make build               - build SAM application"
	@echo "  make validate            - validate SAM template"
	@echo "  make deploy-staging      - deploy to staging"
	@echo "  make deploy-production   - deploy to production"
	@echo ""
	@echo "Legacy aliases:"
	@echo "  make sam-build"
	@echo "  make sam-validate"
	@echo "  make sam-deploy-staging"
	@echo "  make sam-deploy-production"

# =========================
# LOCAL (Docker)
# =========================

up:
	docker compose up -d --build

down:
	docker compose down

docker-build:
	docker compose build

logs:
	docker compose logs -f

restart:
	docker compose down
	docker compose up -d --build

clean:
	docker compose down -v

# =========================
# TESTS
# =========================

test:
	cd treuepunkte-function && go test ./...

test-unit:
	cd treuepunkte-function && go test ./internal/...

test-integration:
	cd treuepunkte-function && go test ./integrationtests/...

# =========================
# SCHEMA PREP
# =========================

prepare-schema-init:
	mkdir -p schema-init/sql
	cp sql/schema/001_schema.sql schema-init/sql/001_schema.sql

# =========================
# SAM (AWS) COMMANDS
# =========================

build: prepare-schema-init
	CGO_ENABLED=0 sam build --no-cached

validate: prepare-schema-init
	sam validate

deploy-staging: prepare-schema-init
	CGO_ENABLED=0 sam build --no-cached
	sam deploy --config-env staging

deploy-production: prepare-schema-init
	CGO_ENABLED=0 sam build --no-cached
	sam deploy --config-env production

# =========================
# LEGACY SAM ALIASES
# =========================

sam-build: build

sam-validate: validate

sam-deploy-staging: deploy-staging

sam-deploy-production: deploy-production