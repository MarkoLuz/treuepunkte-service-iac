# Treuepunkte Service – Project Description for IHK

## 1. Project Title

Treuepunkte Service – Backend System for Managing Customer Loyalty Points

---

## 2. Initial Situation (Ausgangssituation)

In modern e-commerce systems, customer loyalty is essential for long-term success.

Currently, there is no technical system for managing loyalty points. Customers do not receive rewards for repeated purchases, and there is no structured way to track or process such benefits.

This leads to:

- missing incentives for repeat purchases
- high dependency on paid marketing
- lack of transparency of customer benefits

Without a structured loyalty system, it is more difficult to encourage repeat purchases and retain customers.

This increases dependency on paid marketing and can lead to higher customer acquisition costs.

Even small improvements in customer retention can have a significant impact on long-term revenue.

---

## 3. Problem Statement

The main problem is the **lack of a reliable system for handling loyalty points**.

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

→ every action is stored as a new transaction
→ existing data is never overwritten

This ensures:

- complete history
- better debugging
- reliable data handling

---

## 6. Technical Concept

### 6.1 Ledger Principle

I decided to use a ledger model instead of directly updating balances because it allows full traceability of all transactions.

Each action creates a new entry in `points_ledger`, while existing data remains unchanged.

This approach ensures:

- full traceability of all transactions
- clear audit trail
- easier debugging and validation

### 6.2 Design Decision

An alternative approach would be to directly update the customer’s point balance.

However, this would make it difficult to track historical changes and identify the cause of errors.

The ledger approach was chosen because it provides full traceability and allows reconstruction of all transactions at any point in time.

### 6.3 Idempotency

Duplicate requests are common in distributed systems, for example due to network retries.

Without protection, the same request could be processed multiple times, leading to incorrect balances.

To prevent this, idempotency is used so that each request is processed only once, either by returning a conflict response or by safely ignoring duplicate requests without changing the system state.

### 6.4 Database Design

The system uses a relational database (MariaDB) with three main tables:

- `customers`
- `balances`
- `points_ledger`

Concept:

- `points_ledger` stores all events
- `balances` stores aggregated values

This ensures both performance and traceability.

### 6.5 System Architecture

The system is implemented in Go with a layered architecture:

- HTTP layer (API endpoints)
- Service layer (business logic)
- Storage layer (database access)

I decided to use a layered architecture to clearly separate responsibilities and simplify testing.

This separation improves maintainability and testability.

### 6.6 Deployment (AWS)

The system was deployed and tested in AWS using the following services:

- AWS Lambda for executing the backend logic  
- Amazon API Gateway for exposing HTTP endpoints  
- Amazon RDS (MariaDB) for persistent storage  

This setup enables stateless request processing and allows the system to scale according to demand.

### 6.7 Data Model Decision

The `points_ledger` table contains additional fields such as `currency`, `shipping_cents`, `home24_merch_cents` and `mirakl_merch_cents`.

These fields are part of the data model to reflect real e-commerce transaction data and to support traceability.

However, the main focus of this project is the reliable processing of point transactions and the prevention of duplicate bookings.

For this reason, these fields are not described as separate business rules in the project scope.

### 6.8 Example Problem Scenario

In distributed systems, the same request can be sent multiple times, for example due to network retries.

If an order confirmation is processed twice, points could be granted twice.

This would lead to incorrect balances.

This scenario was a key reason why I introduced idempotency, ensuring that each request is processed only once.

---

## 7. Why This Solution

The solution was designed to address specific real-world problems.

Duplicate requests are a common issue in distributed systems.
Without protection, they can lead to incorrect balances.

To solve this, idempotency is used to ensure that each request is processed only once.

Another key requirement was traceability.

Instead of modifying existing data, a ledger model was chosen so that every change is stored as a separate transaction.

This makes it possible to reconstruct the full history of a customer's points at any time.

These decisions ensure that the system is robust, understandable, and suitable for real-world usage.

---

## 8. Example Flow

This flow represents the typical lifecycle of loyalty points in an e-commerce system.

1. Customer places an order
2. System creates an "accrue" transaction (pending points)
3. After confirmation → "confirm" transaction
4. Customer redeems points → "redeem"
5. If needed → "restore" or "revoke"

Each step is stored as a separate entry.

---

## 9. Testing

To ensure correctness and reliability, a multi-level testing approach was used.

### Unit Tests

Unit tests were implemented to validate parts of the business logic independently of the database.

The focus is on input validation and error handling in the service layer.

A simple unit test was implemented to validate input handling in the service layer.

### Integration Tests (Local)

Integration tests were performed using AWS SAM local (`sam local start-api`).

They verify the interaction between:

- HTTP layer
- service layer
- database

All main operations were tested:

- accrue
- confirm
- redeem
- restore
- revoke

The database state was verified after each request.

### End-to-End Tests

Complete business flows were tested:

accrue → confirm → redeem → restore → revoke

This ensures that the full lifecycle behaves correctly.

### Idempotency Tests

Duplicate requests were tested by sending the same request multiple times.

Results:

- first request → processed successfully
- second request → rejected (409 Conflict) or safely ignored
- no duplicate entries in database

This confirms correct handling of duplicate requests.

### Error Scenario Tests

Negative scenarios were tested to validate system robustness:

- redeem without sufficient points
- confirm without prior accrue
- restore without previous redeem
- duplicate operations

The system returned correct HTTP errors and maintained consistent data.

### AWS Integration Tests

The system was also tested in a real AWS environment:

- AWS Lambda
- API Gateway
- Amazon RDS

All endpoints were verified:

- write operations (accrue, confirm, redeem, restore, revoke)
- read operations (balance, transactions)

Additionally, an automated integration test in Go was implemented, which executes requests against the deployed API and verifies responses and database consistency.

---

## 10. Result

The result is a backend system that:

- correctly processes loyalty points
- prevents duplicate transactions
- stores all operations in a traceable way

The system was designed and implemented independently, including architecture, data model, and business logic.

It was successfully validated both locally and in a real cloud environment.

The project demonstrates the ability to design and implement a reliable backend system for real-world scenarios.

---

## 11. Future Improvements

Possible extensions:

- frontend for customers
- reporting and analytics
- extended business rules
- monitoring and logging