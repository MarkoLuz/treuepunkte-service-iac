# Treuepunkte Service

## IHK Version

This repository also includes a version of the README adapted for IHK project requirements.

See [README_IHK.md](./README_IHK.md)

---

## Overview

The Treuepunkte Service is a backend application for managing customer loyalty points in an e-commerce context.

The system processes events related to orders and manages the full lifecycle of loyalty points, including accrual, confirmation, redemption, revocation, and restoration.

The application is implemented in Go and runs as a serverless service on AWS Lambda. Persistent data is stored in a MariaDB database (Amazon RDS in AWS, Docker-based MariaDB locally).

The project demonstrates a complete workflow from local development to cloud deployment using Infrastructure-as-Code (AWS SAM).

---

## Architecture

The system is implemented as a layered backend service with a clear separation of concerns.

### Components

- **AWS Lambda**  
  Executes the application logic.

- **API Gateway (HTTP API)**  
  Exposes REST endpoints and routes requests to Lambda.

- **Amazon RDS (MariaDB)**  
  Stores all transactional data and customer balances.

- **AWS SAM**  
  Defines and deploys infrastructure as code.

- **Docker (local)**  
  Provides a reproducible local development environment.

---

### Internal Structure

The application follows a layered architecture:

- **HTTP Layer**  
  Handles routing, request parsing, and response formatting.

- **Service Layer**  
  Contains business logic and enforces domain rules.

- **Storage Layer**  
  Manages database interaction.

- **Domain Layer**  
  Defines core models and business rules.

---

### Data Model Concept

The system uses an **event-based ledger approach**.

All operations are stored as immutable transactions in a ledger table. This ensures:

- full traceability  
- auditability  
- no loss of historical data  

A separate `balances` table stores the current state for efficient reads.

---

### Idempotency

To prevent duplicate processing, the system uses idempotency keys.

Repeated requests with the same key do not create duplicate transactions.

---

## Features

- Accrue points (`accrue`)
- Confirm points (`confirm`)
- Revoke points (`revoke`)
- Redeem points (`redeem`)
- Restore points (`restore`)
- Retrieve customer balance
- Retrieve transaction history
- Health check endpoint

---

## Tech Stack

- Go (Golang)
- AWS Lambda (provided.al2023)
- Amazon API Gateway (HTTP API)
- Amazon RDS (MariaDB)
- AWS SAM (Infrastructure-as-Code)
- Docker / Docker Compose (local development)
- Makefile (workflow automation)
- Git & GitHub

---

## Project Structure

```text
.
├── Dockerfile
├── Makefile
├── README.md
├── README_IHK.md
├── docker-compose.yml
├── events
│   └── event.json
├── openapi.yaml
├── samconfig.toml
├── schema-init
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   └── sql
│       └── 001_schema.sql   # generated at build time (not versioned)
├── sql
│   └── schema
│       └── 001_schema.sql   # canonical database schema
├── template.yaml
└── treuepunkte-function
    ├── go.mod
    ├── go.sum
    ├── integrationtests
    │   └── aws_integration_test.go
    ├── internal
    │   ├── config
    │   │   └── config.go
    │   ├── domain
    │   │   ├── errors.go
    │   │   ├── models.go
    │   │   └── rules.go
    │   ├── http
    │   │   ├── dto.go
    │   │   ├── errors.go
    │   │   ├── handlers.go
    │   │   └── router.go
    │   ├── service
    │   │   ├── loyalty.go
    │   │   └── loyalty_test.go
    │   └── storage
    │       ├── certs
    │       │   └── global-bundle.pem
    │       ├── mysql.go
    │       └── repo.go
    └── main.go
```

### Database Schema

The database schema is defined in a single canonical file:

- `sql/schema/001_schema.sql`

For local development, this file is mounted into the MariaDB container and executed automatically.

For AWS deployments, the schema is applied by a dedicated Lambda (`schema-init`).  
Because Go's `embed` requires files to be present locally, the schema file is copied into `schema-init/sql/` during the build process.

This file is a generated build artifact and is not version-controlled.

---

## Running the Project

### Prerequisites

- Docker
- Go
- AWS CLI (configured)
- AWS SAM CLI

---

### Local Development (Docker)

Start the local environment:
```bash
make up
```

View logs:
```bash
make logs
```

Stop the environment:
```bash
make down
```

The application will be available at:

`http://localhost:8080`

---

### Tests

Run all tests:

```bash
make test
```

Run unit tests:

```bash
make test-unit
```

Run integration tests:

```bash
make test-integration
```

---

### Deployment (AWS)

Deploy to staging:

```bash
make deploy-staging
```

Deploy to production:

```bash
make deploy-production
```

The deployment process uses AWS SAM and predefined configuration profiles.

---

## CI/CD Pipeline

This project uses a CI/CD pipeline implemented with GitHub Actions and AWS SAM.

The pipeline ensures that every code change is automatically tested, validated, and deployed in a controlled way.

### Continuous Integration (CI)

On every push to the main branch and on every pull request, the following steps are executed:

- Checkout repository
- Set up Go environment
- Run unit and integration tests (`make test`)
- Validate the SAM template (`make validate`)
- Build the application (`make build`)

This guarantees that only working and valid code proceeds to deployment.

### Continuous Deployment – Staging

After a successful CI run, the application is automatically deployed to the staging environment.

**Trigger:** `git push → main`

**Deployment:** `sam deploy --config-env staging`

This allows immediate testing of changes in a cloud environment without manual intervention.

### Continuous Deployment – Production

Deployment to production is intentionally manual to ensure safety and control.

**Trigger:** GitHub → Actions → CI/CD → Run workflow

**Input:** `deploy_production = yes`

**Deployment:** `sam deploy --config-env production`

This prevents accidental deployments and follows best practices for controlled releases.

### Security

AWS credentials are not stored in the codebase.
They are securely managed using GitHub repository secrets:

- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`

### Summary

The pipeline follows a standard and production-ready workflow:

- Automated testing and validation (CI)
- Automatic deployment to staging
- Manual, controlled deployment to production

This setup ensures reliability, reproducibility, and alignment with Infrastructure-as-Code principles.

---

## Database Initialization

The database schema is initialized automatically, depending on the environment:

**Local (Docker)**  
MariaDB executes SQL scripts from `sql/init/` via `docker-entrypoint-initdb.d`.

**AWS**  
A dedicated Lambda function (`schema-init`) initializes the schema during deployment using a CloudFormation custom resource.

This ensures that the required tables are created automatically in both environments.

---

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/v1/points/accrue` | Create pending points |
| POST | `/v1/points/confirm` | Confirm points |
| POST | `/v1/points/revoke` | Revoke points |
| POST | `/v1/points/redeem` | Redeem points |
| POST | `/v1/points/restore` | Restore points |
| GET | `/v1/customers/{id}/balance` | Get current balance |
| GET | `/v1/customers/{id}/transactions` | Get transaction history |

---

## Configuration

The application is configured via environment variables.

### Main variables

- APP_ENV
- APP_PORT
- DB_HOST
- DB_PORT
- DB_USER
- DB_PASS
- DB_NAME

### Local Environment

Docker Compose provides default values:

```env
DB_USER=treuepunkte
DB_PASS=treuepunkte
DB_NAME=treuepunkte
```

These values are for development only.

### AWS Environment

- Database credentials are stored in AWS Secrets Manager
- Password is injected via CloudFormation dynamic reference
- No secrets are stored in the repository

---

## Testing Strategy

The project includes:

- Unit tests (service layer)
- Integration tests (API + database interaction)
- Manual end-to-end tests (via HTTP requests)

Test scenarios include:

- full transaction flows
- error handling
- idempotency behavior

---

## Notes / Limitations

- Local and AWS database configurations use slightly different naming conventions
- Database schema is currently duplicated for Docker and AWS initialization

These limitations are known and do not affect the correctness of the system, but are candidates for future refactoring.

---

## Future Improvements

- Unify database schema source (single source of truth)
- Improve test coverage
- Add structured logging and monitoring
- Add secret rotation via AWS Secrets Manager


