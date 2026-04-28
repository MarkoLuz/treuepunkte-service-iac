
# Treuepunkte Service – Project Description for IHK

## 1. Project Title

**Treuepunkte Service – Backend System for Managing Customer Loyalty Points**

---

## 2. Initial Situation (Ausgangssituation)

In modern e-commerce systems, customer loyalty is essential for long-term success.

Currently, there is no technical system for managing loyalty points. Customers do not receive rewards for repeated purchases, and there is no structured way to track or process such benefits.

This leads to:

- missing incentives for repeat purchases
- limited customer retention capabilities
- lack of structured reward system

Without a structured loyalty system, it is more difficult to encourage repeat purchases and retain customers.  
This increases dependency on paid marketing and can lead to higher customer acquisition costs.

Even small improvements in customer retention can have a significant impact on long-term revenue.

---

## 3. Problem Statement

The main problem is the lack of a reliable system for handling loyalty points.

Without such a system:

- duplicate requests can lead to incorrect balances
- there is no traceable history of transactions
- manual corrections may be required

The system must ensure:

- correct processing of all transactions
- prevention of duplicate execution
- consistent balances
- full traceability

---

## 4. Project Objective (Ziel des Projekts)

The goal of this project is to develop a backend service that:

- manages loyalty points automatically
- ensures consistent and correct processing
- stores all transactions in a traceable way
- provides a stable API for integration

The system focuses on the lifecycle of points:

- accrue
- confirm
- redeem
- restore
- revoke

---

## 5. Solution Overview

The solution is a backend service that processes all point-related events.

An event-based approach is used:

- every action is stored as a new transaction
- existing data is never overwritten

This ensures:

- complete history
- better debugging
- reliable data handling

---

## 6. Technical Concept

### 6.1 Ledger Principle

A ledger model is used instead of directly updating balances, allowing full traceability of all transactions.

Each action creates a new entry in `points_ledger`, while existing data remains unchanged.

### 6.2 Idempotency

Duplicate requests are common in distributed systems (e.g., network retries).

Idempotency ensures each request is processed only once and prevents duplicate transactions.

### 6.3 Database Design

The system uses a relational database (MariaDB) with three main tables:

- `customers`
- `balances`
- `points_ledger`

**Concept:**

- `points_ledger` stores all events
- `balances` stores aggregated values

### 6.4 System Architecture

The system is implemented in Go with a layered architecture:

- HTTP layer
- service layer
- storage layer

This separation improves maintainability and testability.

### 6.5 Deployment (AWS)

The system is deployed using AWS SAM (Infrastructure-as-Code).

The infrastructure can be created and removed at any time using the same configuration.

**Services used:**

- AWS Lambda (backend execution)
- API Gateway (HTTP API)
- Amazon RDS (MariaDB)

**Environments:**

- staging
- production

Deployment is executed via predefined Makefile commands.

A CI/CD pipeline is implemented using GitHub Actions:

- automatic deployment to staging on push to main
- manual deployment to production

This ensures controlled and reproducible releases.

### 6.6 Database Initialization

The database schema is initialized automatically depending on the environment.

- In the local environment, MariaDB executes SQL scripts via Docker (`docker-entrypoint-initdb.d`)
- In AWS, a dedicated Lambda function (`schema-init`) initializes the schema during deployment as part of the CloudFormation lifecycle

This ensures that the database structure is created consistently in both environments.

---

## 7. Example Flow

1. Customer places an order
2. System creates an "accrue" transaction (pending)
3. Order is confirmed → "confirm"
4. Customer redeems points → "redeem"
5. If needed → "restore" or "revoke"

Each step is stored as a separate transaction.

---

## 8. Testing

A multi-level testing strategy was used.

The system behaved consistently across all environments.

### Unit Tests

- Validate business logic independently of the database.

### Integration Tests (Local)

Integration tests are executed against the application and database using the local Docker environment.

They verify the interaction between:

- HTTP layer
- service layer
- database

### End-to-End Tests

Complete flows were tested:

- accrue → confirm → redeem → restore → revoke

### Idempotency Tests

Duplicate requests were tested to ensure:

- no duplicate transactions
- consistent state

### AWS Tests

The system was tested in a real AWS environment:

- Lambda
- API Gateway
- RDS

All endpoints were verified successfully.

---

## 9. Result

The result is a backend system that:

- correctly processes loyalty points
- prevents duplicate transactions
- stores all operations in a traceable way
- can be deployed and reproduced using Infrastructure-as-Code

The system was successfully validated both locally and in AWS.

---

## 10. Future Improvements

- monitoring and logging
- improved test coverage
- secret rotation using AWS Secrets Manager