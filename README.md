# Treuepunkte Service

## Project background

This project was developed in the context of my internship at home24.

The business case is based on an e-commerce loyalty-points scenario with home24 merchandise and Mirakl/marketplace order items. The service calculates loyalty points from order amounts, stores the transaction history in a ledger table and maintains the current customer balance separately.

For the publicly documented and independently reproducible version of the project, internal company infrastructure is not used. Instead, the AWS deployment is implemented and validated in a private AWS account using an Infrastructure as Code approach.

This keeps the public project version clearly separated from internal company systems, while still demonstrating how the application can be deployed and validated in a real AWS environment.

---

## What this project demonstrates

This project is not a generic CRUD API. It focuses on a realistic loyalty-points lifecycle with transactional consistency between a ledger table and a current-balance table.

Key parts:

- Go backend service with HTTP API
- loyalty lifecycle: accrue, confirm, revoke, redeem, restore
- point calculation based on order amounts
- ledger-based transaction history
- current balance read model
- idempotency and duplicate protection
- MariaDB schema for local and AWS environments
- local Docker Compose setup
- AWS SAM infrastructure for Lambda, API Gateway, RDS and VPC networking
- GitHub Actions CI with manual AWS deployments

---

## Business rules

Points are calculated by the service during accrual. The client does not send the final number of points for an accrual request.

Current accrual rules:

| Order amount type                | Rule              |
| -------------------------------- | ----------------- |
| home24 merchandise               | 10 points per EUR |
| Mirakl / marketplace merchandise | 5 points per EUR  |
| Shipping                         | 0 points          |
| Currency                         | EUR only          |

Example:

```text
home24_merch_cents = 1200   -> 120 points
mirakl_merch_cents = 1000   ->  50 points
shipping_cents     =  490   ->   0 points

calculated points = 170
order_total_cents = 2200
```

`order_total_cents` contains only merchandise amounts:

```text
order_total_cents = home24_merch_cents + mirakl_merch_cents
```

Shipping is stored for traceability, but it does not earn points and is not included in `order_total_cents`.

---

## Loyalty lifecycle

The service supports the following operations:

| Operation | Purpose                                               |
| --------- | ----------------------------------------------------- |
| `accrue`  | Create pending points from an order                   |
| `confirm` | Move previously accrued points from pending to active |
| `revoke`  | Remove points when an order is returned               |
| `redeem`  | Spend active points                                   |
| `restore` | Restore previously redeemed points                    |

The service models the waiting period through `pending` and `active` points. The automatic time-based trigger, for example after a return period, is intentionally outside the current scope and can be handled by an external scheduler or an existing company workflow.

The lifecycle is stored in a ledger-style transaction history. Existing ledger entries are not overwritten when the lifecycle changes. Instead, follow-up actions are recorded as separate entries.

For example, confirming an accrual does not update the original `accrue` row. It creates a separate `confirm` ledger entry and updates the customer balance.

---

## Architecture

The application follows a layered backend structure.

```text
HTTP layer
  - routes requests
  - decodes JSON request bodies
  - resolves idempotency keys from body or header
  - maps domain errors to HTTP status codes
  - formats JSON responses

Service layer
  - validates input
  - enforces EUR-only accruals
  - calculates points from order amounts
  - orchestrates loyalty operations

Repository / storage layer
  - executes SQL queries
  - wraps lifecycle operations in database transactions
  - inserts ledger entries
  - updates current balances
  - handles duplicate/idempotency conflicts
  - keeps ledger and balance changes consistent

Domain layer
  - defines shared domain models and errors
```

---

## Data model

The database schema is defined in:

```text
sql/schema/001_schema.sql
```

Main tables:

| Table           | Purpose                                        |
| --------------- | ---------------------------------------------- |
| `customers`     | Stores customer identifiers                    |
| `balances`      | Stores current active and pending point totals |
| `points_ledger` | Stores the loyalty transaction history         |

### Ledger and balance model

The `points_ledger` table is the historical source of truth for loyalty operations. It records entries such as:

- `accrue`
- `confirm`
- `revoke`
- `redeem`
- `restore`

The `balances` table is a read model for efficient balance queries:

```text
active_points
pending_points
```

This avoids recalculating the current balance from the full ledger history on every balance request.

---

## Idempotency and duplicate protection

Write operations support idempotency keys.

An idempotency key can be sent either:

- in the JSON request body as `idempotency_key`
- in the `Idempotency-Key` HTTP header

If both are provided, the request body value is used.

The database schema contains a unique constraint on `idempotency_key`. Additional unique constraints protect business references such as order IDs, redeem references and return IDs.

Repeated or conflicting requests return `409 Conflict`; the current implementation does not replay the original response as `200 OK`.

---

## API endpoints

| Method | Endpoint                                   | Description                                     |
| ------ | ------------------------------------------ | ----------------------------------------------- |
| `GET`  | `/health`                                  | Technical health check, returns plain text `ok` |
| `POST` | `/v1/points/accrue`                        | Create pending points from order amounts        |
| `POST` | `/v1/points/confirm`                       | Confirm previously accrued points               |
| `POST` | `/v1/points/revoke`                        | Revoke points for a returned order              |
| `POST` | `/v1/points/redeem`                        | Redeem active points                            |
| `POST` | `/v1/points/restore`                       | Restore previously redeemed points              |
| `GET`  | `/v1/customers/{customer_id}/balance`      | Get current customer balance                    |
| `GET`  | `/v1/customers/{customer_id}/transactions` | Get customer transaction history                |

The OpenAPI specification is available in:

```text
openapi.yaml
```

---

## Example accrual request

```json
{
  "customer_id": "cust-1",
  "order_id": "order-1001",
  "home24_merch_cents": 1200,
  "mirakl_merch_cents": 1000,
  "shipping_cents": 490,
  "currency": "EUR",
  "idempotency_key": "accrue-1001"
}
```

This creates a pending accrual of 170 points:

```text
1200 cents home24 merchandise -> 120 points
1000 cents Mirakl merchandise ->  50 points
490 cents shipping            ->   0 points
```

---

## Tech stack

- Go
- MariaDB
- Docker / Docker Compose
- AWS Lambda
- Amazon API Gateway
- Amazon RDS for MariaDB
- AWS Secrets Manager
- AWS SAM
- GitHub Actions
- Makefile

---

## Development and validation tools

The project was developed and validated using VS Code, Git/GitHub, Docker Compose, Make, DBeaver, Postman, curl, OpenAPI/Swagger tooling, AWS CLI, AWS SAM CLI and GitHub Actions.

DBeaver was used to inspect the MariaDB schema and test data during local development. Postman and curl were used for manual API and lifecycle testing, primarily against the local Docker environment. The deployed AWS API was validated separately where applicable. The OpenAPI specification documents the REST API contract in `openapi.yaml`.


---

## Project structure

```text
.
├── .github/
│   └── workflows/
│       └── ci.yml
├── Dockerfile
├── Makefile
├── README.md
├── docker-compose.yml
├── events/
│   └── event.json
├── openapi.yaml
├── samconfig.toml
├── schema-init/
│   ├── go.mod
│   ├── go.sum
│   └── main.go
├── sql/
│   └── schema/
│       └── 001_schema.sql
├── template.yaml
└── treuepunkte-function/
    ├── Makefile
    ├── go.mod
    ├── go.sum
    ├── main.go
    ├── integrationtests/
    │   └── aws_integration_test.go
    └── internal/
        ├── config/
        ├── domain/
        ├── http/
        ├── service/
        └── storage/
```

Notes:

- `sql/schema/001_schema.sql` is the canonical database schema.
- `schema-init/` contains the Lambda custom resource code for applying the schema during AWS deployment.
- Build artifacts such as `bootstrap` are not part of the source structure.

---

## Local development

### Prerequisites

- Docker
- Docker Compose
- Go
- Make

### Start local environment

```bash
make up
```

This starts:

- MariaDB container
- Go API container

The local database is initialized from:

```text
sql/schema/001_schema.sql
```

The API is available at:

```text
http://localhost:8080
```

### Health check

```bash
curl http://localhost:8080/health
```

Expected response:

```text
ok
```

### View logs

```bash
make logs
```

### Stop local environment

```bash
make down
```

### Remove local containers and database volume

```bash
make clean
```

Warning: `make clean` removes the Docker volume and deletes local database data.

---

## Testing

Run all Go tests:

```bash
make test
```

Run service/domain tests:

```bash
make test-unit
```

Run integration tests:

```bash
make test-integration
```

The project also includes manual API lifecycle testing with real HTTP requests against the local Docker environment.

Tested scenarios include:

- accrue -> confirm -> redeem -> restore
- revoke before confirm
- revoke after confirm
- duplicate requests with the same idempotency key
- redeeming more points than available
- missing accrue transaction for confirm/revoke
- missing redeem transaction for restore
- validation of invalid or negative input values
- balance consistency after rejected requests
- point calculation for home24 merchandise, Mirakl/marketplace merchandise and shipping
- EUR-only validation for accruals
- rejection of shipping-only accruals

The AWS integration test is opt-in and requires an already deployed API endpoint. It is enabled with `RUN_AWS_INTEGRATION=1` and `AWS_BASE_URL`.

---

## AWS deployment

The AWS infrastructure is defined in:

```text
template.yaml
samconfig.toml
```

The SAM template provisions:

- API Gateway HTTP API
- Go Lambda function
- schema initialization Lambda
- Lambda-backed CloudFormation custom resource
- private RDS MariaDB database
- VPC with public and private subnets
- NAT Gateway
- S3 Gateway VPC Endpoint
- Lambda and database security groups
- Secrets Manager managed database password

The Go Lambda runs inside private subnets and connects to the private RDS instance through the database security group.

The database schema is initialized during deployment by the schema-init Lambda custom resource.

### Validate and build

```bash
make validate
make build
```

### Deploy staging manually

```bash
make deploy-staging
```

Equivalent SAM command:

```bash
sam deploy --config-env staging
```

### Deploy production manually

```bash
make deploy-production
```

Equivalent SAM command:

```bash
sam deploy --config-env production
```

Staging and production use separate SAM configuration environments and separate stack names:

| Environment | Stack name                   |
| ----------- | ---------------------------- |
| staging     | `treuepunkte-iac-staging`    |
| production  | `treuepunkte-iac-production` |

---

## AWS cleanup

AWS deployments create cost-relevant resources, including RDS, NAT Gateway and networking resources.

To delete the staging stack:

```bash
aws cloudformation delete-stack \
  --stack-name treuepunkte-iac-staging \
  --region eu-west-1
```

Wait until deletion is complete:

```bash
aws cloudformation wait stack-delete-complete \
  --stack-name treuepunkte-iac-staging \
  --region eu-west-1
```

Because the RDS resource uses snapshot policies, stack deletion may create a final database snapshot.

---

## CI/CD

GitHub Actions is used for continuous integration and manual deployments.

Workflow file:

```text
.github/workflows/ci.yml
```

### Continuous Integration

CI runs automatically on:

- push to `main`
- pull request to `main`

The CI job runs:

```bash
make test
make validate
make build
```

### Manual staging deployment

Staging deployment is started manually through GitHub Actions by selecting:

```text
deploy_target = staging
```

### Manual production deployment

Production deployment is started manually through GitHub Actions by selecting:

```text
deploy_target = production
confirm_production = yes
```

The production environment is additionally protected by a GitHub Environment required reviewer rule.

Pushes to `main` do not automatically deploy AWS resources.

---

## Configuration

The application is configured through environment variables.

Main variables:

```text
APP_ENV
APP_PORT
DB_HOST
DB_PORT
DB_USER
DB_PASS
DB_NAME
```

### Local configuration

Docker Compose provides local development values:

```text
APP_ENV=local
APP_PORT=8080
DB_HOST=db
DB_PORT=3306
DB_USER=treuepunkte
DB_PASS=treuepunkte
DB_NAME=treuepunkte
```

### AWS configuration

In AWS, database connection values are provided by CloudFormation.

The RDS master password is managed by AWS Secrets Manager and is not stored in the repository.

---

## Current limitations

The current scope focuses on the core loyalty lifecycle and infrastructure deployment.

Not included in the current implementation:

- automatic 14-day activation scheduler
- admin/reporting API
- asynchronous event processing
- RDS Proxy
- CloudWatch alarms
- production-grade observability setup
- database migration tool

---

## Future improvements

Possible next steps:

- add a scheduler for automatic point activation after the waiting period
- introduce a database migration tool
- add RDS Proxy for improved Lambda-to-RDS connection handling
- add CloudWatch alarms and structured operational dashboards
- add API Gateway throttling and rate limiting
- introduce SQS or EventBridge for asynchronous event processing
- add an admin/reporting API
- improve test coverage for repository and HTTP handler edge cases
