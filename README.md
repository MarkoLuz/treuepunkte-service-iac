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

- AWS Lambda
  Executes the application logic.

- API Gateway (HTTP API)
  Exposes REST endpoints and routes requests to Lambda.

- Amazon RDS (MariaDB)
  Stores all transactional data and customer balances.

- AWS SAM
  Defines and deploys infrastructure as code.

- Docker (local)
  Provides a reproducible local development environment.

---

### Internal Structure

The application follows a layered architecture:

- HTTP Layer
  Handles routing, request parsing, and response formatting.

- Service Layer
  Contains business logic and enforces domain rules.

- Storage Layer
  Manages database interaction.

- Domain Layer
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

- accrue points (`accrue`)
- confirm points (`confirm`)
- revoke points (`revoke`)
- redeem points (`redeem`)
- restore points (`restore`)
- retrieve customer balance
- retrieve transaction history
- health check endpoint

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
│   ├── Makefile
│   ├── bootstrap
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── sql
│   └── schema
│       └── 001_schema.sql
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

---

## Database Schema

The database schema is defined in:

- `sql/schema/001_schema.sql`

For local development, this file is mounted into the MariaDB container and executed automatically.

For AWS deployments, the schema is applied by a dedicated Lambda (`schema-init`).

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

```
http://localhost:8080
```

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

---

## Deploy & Run (from scratch)

### Deploy to staging

```bash
sam deploy --config-env staging
```

### Get API endpoint

After deployment, the API endpoint is printed in the output.

Set it as an environment variable:

```bash
export API_URL="https://<api-id>.execute-api.eu-west-1.amazonaws.com"
```

### Test the service

```bash
curl "$API_URL/health"
```

---

## Cleanup

To remove all AWS resources (staging):

```bash
aws cloudformation delete-stack \
  --stack-name treuepunkte-iac-staging
```

---

## CI/CD Pipeline

This project uses GitHub Actions for CI/CD.

### Continuous Integration (CI)

On every push to the main branch and on every pull request:

- run tests (`make test`)
- validate SAM template (`make validate`)
- build application (`make build`)

### Continuous Deployment – Staging

Automatic deployment on push to main:

```bash
sam deploy --config-env staging
```

### Continuous Deployment – Production

Manual deployment via GitHub Actions.

---

## Database Initialization

**Local (Docker)**
MariaDB executes SQL scripts from `sql/schema/`.

**AWS**
A dedicated Lambda (`schema-init`) initializes the schema during deployment.

---

## API Endpoints

| Method | Endpoint                          | Description             |
| ------ | --------------------------------- | ----------------------- |
| GET    | `/health`                         | Health check            |
| POST   | `/v1/points/accrue`               | Create pending points   |
| POST   | `/v1/points/confirm`              | Confirm points          |
| POST   | `/v1/points/revoke`               | Revoke points           |
| POST   | `/v1/points/redeem`               | Redeem points           |
| POST   | `/v1/points/restore`              | Restore points          |
| GET    | `/v1/customers/{id}/balance`      | Get current balance     |
| GET    | `/v1/customers/{id}/transactions` | Get transaction history |

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

### AWS Environment

- credentials stored in AWS Secrets Manager
- injected via CloudFormation
- no secrets in the repository

---

## Testing Strategy

- Unit tests (service layer)
- Integration tests (API + database interaction)
- Manual end-to-end tests

---

## Notes / Limitations

- Test coverage is focused on core functionality

---

## Future Improvements

- Improve test coverage
- Add structured logging and monitoring
- Implement secret rotation
