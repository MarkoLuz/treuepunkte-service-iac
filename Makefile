.PHONY: help
.PHONY: up down docker-build logs restart clean
.PHONY: test test-unit test-integration
.PHONY: build validate deploy-staging deploy-production
.PHONY: sam-build sam-validate sam-deploy-staging sam-deploy-production
.PHONY: build-SchemaInitFunction

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
# SAM CUSTOM BUILD TARGETS
# =========================

build-SchemaInitFunction:
	@set -e; \
	cp sql/schema/001_schema.sql schema-init/001_schema.sql; \
	trap 'rm -f schema-init/001_schema.sql schema-init/bootstrap' EXIT; \
	( cd schema-init && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go ); \
	chmod +x schema-init/bootstrap; \
	mkdir -p $(ARTIFACTS_DIR); \
	cp schema-init/bootstrap $(ARTIFACTS_DIR)/bootstrap

# =========================
# SAM (AWS) COMMANDS
# =========================

build:
	CGO_ENABLED=0 sam build --no-cached

validate:
	sam validate

deploy-staging:
	CGO_ENABLED=0 sam build --no-cached
	sam deploy --config-env staging

deploy-production:
	CGO_ENABLED=0 sam build --no-cached
	sam deploy --config-env production

# =========================
# LEGACY SAM ALIASES
# =========================

sam-build: build

sam-validate: validate

sam-deploy-staging: deploy-staging

sam-deploy-production: deploy-production