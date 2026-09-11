# Payment Platform

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge\&logo=go\&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-Framework-00ADD8?style=for-the-badge\&logo=gin\&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-4169E1?style=for-the-badge\&logo=postgresql\&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-Rate%20Limiting-DC382D?style=for-the-badge\&logo=redis\&logoColor=white)
![Kafka](https://img.shields.io/badge/Kafka-Event%20Streaming-231F20?style=for-the-badge\&logo=apachekafka\&logoColor=white)
![JWT](https://img.shields.io/badge/JWT-Authentication-000000?style=for-the-badge\&logo=jsonwebtokens\&logoColor=white)

A Go microservices payment platform: an API gateway authenticates and rate-limits requests, an auth service issues JWTs, a payment service owns customers, payments, and refunds, and a notification service asynchronously sends email notifications through Kafka.

> **Portfolio / learning project:** This project simulates a payment platform and does not connect to a real payment processor or bank. Payment success/failure is currently triggered through development-only endpoints that simulate provider callbacks.

## Table of Contents

* [Features](#features)
* [Tech Stack](#tech-stack)
* [Architecture](#architecture)
* [Services](#services)
* [Project Structure](#project-structure)
* [Authentication](#authentication)
* [Payments & Refunds](#payments--refunds)
* [Payment Events](#payment-events)
* [API Endpoints](#api-endpoints)
* [Data Model](#data-model)
* [Setup](#setup)
* [Security](#security)
* [Known Limitations](#known-limitations)
* [Future Improvements](#future-improvements)
* [License](#license)

## Features

* Registration and login with bcrypt and JWT authentication
* API gateway as the intended public entry point
* Centralized JWT verification with `X-Merchant-ID` identity propagation
* Redis-backed per-merchant, per-route rate limiting
* Idempotent payment creation using `Idempotency-Key`
* Merchant-scoped customers, payments, and refunds
* Payment state machine with validated transitions
* Full and partial refunds
* Kafka-based asynchronous payment events
* SMTP email notifications
* Separate PostgreSQL databases for auth and payment services
* Layered handler → service → repository architecture

## Tech Stack

* **Language:** Go 1.26
* **Web Framework:** [Gin](https://github.com/gin-gonic/gin)
* **Database:** PostgreSQL via [pgx/v5](https://github.com/jackc/pgx)
* **Cache / Rate Limiting:** Redis via [go-redis/v9](https://github.com/redis/go-redis)
* **Messaging:** Kafka via [segmentio/kafka-go](https://github.com/segmentio/kafka-go)
* **Authentication:** [golang-jwt/jwt](https://github.com/golang-jwt/jwt) + bcrypt
* **Email:** Go `net/smtp`
* **Architecture:** Microservices + API Gateway + event-driven notifications

## Architecture

```mermaid
flowchart LR
    Client[Client]

    subgraph Gateway["api-gateway :8080"]
        MW["JWT auth → Idempotency-Key check → Rate limiter"]
    end

    Auth["auth-service :8081<br/>register, login"]
    Payment["payment-service :8082<br/>customers, payments, refunds"]
    Notification["notification-service<br/>Kafka consumer"]

    Redis[(Redis)]
    AuthDB[(PostgreSQL<br/>auth_db)]
    PaymentDB[(PostgreSQL<br/>payment_db)]
    Kafka[/Kafka<br/>payment.events/]
    SMTP[SMTP Server]

    Client --> MW
    MW -->|auth routes| Auth
    MW -->|protected routes| Payment
    MW <--> Redis

    Auth --> AuthDB
    Payment --> PaymentDB
    Payment -->|publish| Kafka
    Kafka -->|consume| Notification
    Notification --> SMTP
```

The **API Gateway is the only intended public application entry point**. It verifies JWTs, validates idempotency keys, applies rate limits, and forwards the authenticated user's ID through `X-Merchant-ID`.

`payment-service` trusts this header instead of independently verifying the JWT. Therefore, direct access to `payment-service` can bypass the gateway's authentication boundary.

## Services

| Service                |   Port | Responsibility                                                            |
| ---------------------- | -----: | ------------------------------------------------------------------------- |
| `api-gateway`          | `8080` | JWT verification, rate limiting, idempotency validation, request proxying |
| `auth-service`         | `8081` | Registration, login, password hashing, JWT issuance                       |
| `payment-service`      | `8082` | Customers, payments, processing, refunds, Kafka events                    |
| `notification-service` |      — | Kafka consumer and SMTP notifications                                     |

Each service is an independent Go module.

## Project Structure

```text
Payment-Platform/
├── api-gateway/
│   ├── cmd/server/
│   └── internal/
│       ├── config/
│       ├── middlewares/
│       └── proxy/
│
├── auth-service/
│   ├── cmd/server/
│   ├── internal/
│   │   ├── config/
│   │   ├── db/
│   │   ├── handlers/
│   │   ├── jwt/
│   │   ├── models/
│   │   ├── password/
│   │   ├── repository/
│   │   ├── services/
│   │   └── validation/
│   └── migrations/
│
├── payment-service/
│   ├── cmd/server/
│   ├── internal/
│   │   ├── config/
│   │   ├── db/
│   │   ├── handlers/
│   │   ├── kafka/
│   │   ├── models/
│   │   ├── repository/
│   │   ├── services/
│   │   ├── utils/
│   │   └── validations/
│   └── migrations/
│
└── notification-service/
    ├── cmd/server/
    └── internal/
        ├── config/
        ├── kafka/
        ├── mail/
        ├── models/
        └── services/
```

## Authentication

`auth-service` stores users, while the payment domain uses the authenticated user's ID as `merchant_id`. There is no separate merchant entity.

### Registration

```text
Client → api-gateway → auth-service → PostgreSQL
```

Passwords are bcrypt-hashed before storage.

**Request:**

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "at-least-8-chars"
}
```

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Jane Doe",
  "email": "jane@example.com",
  "created_at": "2026-09-11T20:15:00Z",
  "updated_at": "2026-09-11T20:15:00Z"
}
```

### Login

```text
Client → api-gateway → auth-service
                         ↓
                       JWT
```

The JWT:

* Contains the user ID in the `id` claim
* Expires after 8 hours
* Uses HMAC signing
* Has no refresh-token or server-side revocation mechanism

**Request:**

```json
{
  "email": "jane@example.com",
  "password": "your-password"
}
```

**Response:**

```json
{
  "token": "<jwt>"
}
```

## Payments & Refunds

### Payment Lifecycle

```mermaid
stateDiagram-v2
    [*] --> pending
    pending --> processing
    processing --> succeeded
    processing --> failed
```

Supported currencies:

```text
USD
EUR
TRY
```

### Idempotency

Payment creation requires an `Idempotency-Key`:

```http
POST /payments
Authorization: Bearer <JWT>
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
```

The database enforces:

```text
UNIQUE(merchant_id, idempotency_key)
```

Retrying the same payment with the same key returns the existing payment instead of creating a duplicate.

### Refunds

Refunds are supported for `succeeded` and `partially_refunded` payments.

```text
succeeded
    ↓
partially_refunded
    ↓
refunded
```

Multiple partial refunds can be issued until the payment is fully refunded. Refund creation and the payment's cumulative refunded amount are updated in the same PostgreSQL transaction.

## Payment Events

`payment-service` publishes the following events to the `payment.events` Kafka topic:

* `payment.created`
* `payment.processing`
* `payment.succeeded`
* `payment.failed`
* `payment.refunded`
* `payment.partially_refunded`

```text
payment-service → Kafka → notification-service → SMTP
```

`notification-service` consumes the events and sends event-specific emails asynchronously.

## API Endpoints

All normal API routes are exposed through:

```text
http://localhost:8080
```

### Authentication

| Method | Path             | Auth |
| ------ | ---------------- | ---- |
| POST   | `/auth/register` | —    |
| POST   | `/auth/login`    | —    |

**POST `/auth/register`**

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "at-least-8-chars"
}
```

**Response:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Jane Doe",
  "email": "jane@example.com",
  "created_at": "2026-09-11T20:15:00Z",
  "updated_at": "2026-09-11T20:15:00Z"
}
```

**POST `/auth/login`**

```json
{
  "email": "jane@example.com",
  "password": "your-password"
}
```

**Response:**

```json
{
  "token": "<jwt>"
}
```

### Customers

| Method | Path             | Rate Limit |
| ------ | ---------------- | ---------: |
| POST   | `/customers`     |     20/min |
| GET    | `/customers`     |     60/min |
| GET    | `/customers/:id` |     60/min |

**POST `/customers`**

```json
{
  "name": "John Smith",
  "email": "john@example.com"
}
```

**Response:**

```json
{
  "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "John Smith",
  "email": "john@example.com",
  "created_at": "2026-09-11T20:20:00Z"
}
```

### Payments

| Method | Path                    | Rate Limit |
| ------ | ----------------------- | ---------: |
| POST   | `/payments`             |     20/min |
| GET    | `/payments`             |     60/min |
| GET    | `/payments/:id`         |     60/min |
| POST   | `/payments/:id/process` |     20/min |

**POST `/payments`**

Headers:

```http
Authorization: Bearer <JWT>
Idempotency-Key: <uuid>
```

Body:

```json
{
  "customer_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "amount_cents": 5000,
  "currency": "USD"
}
```

**Response:**

```json
{
  "id": "7f3c9f4e-4c4a-4d6a-8c8a-123456789abc",
  "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
  "customer_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "amount_cents": 5000,
  "refunded_amount_cents": 0,
  "currency": "USD",
  "status": "pending",
  "idempotency_key": "550e8400-e29b-41d4-a716-446655440001",
  "created_at": "2026-09-11T20:25:00Z",
  "updated_at": "2026-09-11T20:25:00Z"
}
```

### Refunds

| Method | Path                    | Rate Limit |
| ------ | ----------------------- | ---------: |
| POST   | `/payments/:id/refunds` |     10/min |
| GET    | `/payments/:id/refunds` |     60/min |

**POST `/payments/:id/refunds`**

```json
{
  "amount_cents": 2000
}
```

**Response:**

```json
{
  "id": "8ac4c2b7-3e7f-4f1a-b6c2-987654321def",
  "merchant_id": "550e8400-e29b-41d4-a716-446655440000",
  "customer_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "amount_cents": 5000,
  "refunded_amount_cents": 2000,
  "currency": "USD",
  "status": "partially_refunded",
  "idempotency_key": "550e8400-e29b-41d4-a716-446655440001",
  "created_at": "2026-09-11T20:25:00Z",
  "updated_at": "2026-09-11T20:30:00Z"
}
```

### Development-Only Payment Simulation

These endpoints simulate callbacks from a payment provider:

| Method | Path                    |
| ------ | ----------------------- |
| POST   | `/payments/:id/succeed` |
| POST   | `/payments/:id/fail`    |

They are registered directly on `payment-service`, so they bypass gateway JWT verification and rate limiting.

They still require:

```http
X-Merchant-ID: <merchant-uuid>
```

These endpoints should be removed when real provider webhooks are implemented.

## Data Model

```mermaid
erDiagram
    USERS {
        uuid id PK
        text name
        text email UK
        text password_hash
        timestamptz created_at
        timestamptz updated_at
    }

    CUSTOMERS {
        uuid id PK
        uuid merchant_id
        text name
        text email
        timestamptz created_at
    }

    PAYMENTS {
        uuid id PK
        uuid merchant_id
        uuid customer_id FK
        bigint amount_cents
        bigint refunded_amount_cents
        varchar currency
        varchar status
        uuid idempotency_key
        timestamptz created_at
        timestamptz updated_at
    }

    REFUNDS {
        uuid id PK
        uuid payment_id FK
        uuid merchant_id
        bigint amount_cents
        timestamptz created_at
    }

    CUSTOMERS ||--o{ PAYMENTS : has
    PAYMENTS ||--o{ REFUNDS : has
```

`auth-service` owns `auth_db`.

`payment-service` owns `payment_db`.

There are no cross-service database foreign keys. `merchant_id` is the authenticated user's ID propagated through `X-Merchant-ID`.

## Setup

### Prerequisites

* Go 1.26+
* PostgreSQL
* Docker
* Redis
* Kafka
* SMTP account/relay

### Clone

```bash
git clone https://github.com/ErenKarakus1/Payment-Platform.git
cd Payment-Platform
```

### Environment Variables

Each service contains a `.env.example`.

**Bash / macOS / Linux:**

```bash
cp api-gateway/.env.example api-gateway/.env
cp auth-service/.env.example auth-service/.env
cp payment-service/.env.example payment-service/.env
cp notification-service/.env.example notification-service/.env
```

**PowerShell:**

```powershell
Copy-Item api-gateway/.env.example api-gateway/.env
Copy-Item auth-service/.env.example auth-service/.env
Copy-Item payment-service/.env.example payment-service/.env
Copy-Item notification-service/.env.example notification-service/.env
```

Set the required values in each `.env`.

`JWT_SECRET` must be identical in `api-gateway` and `auth-service`.

### PostgreSQL

Create the databases:

```sql
CREATE DATABASE auth_db;
CREATE DATABASE payment_db;
```

Apply the migrations:

```bash
psql "postgres://postgres:your_password@localhost:5432/auth_db" -f auth-service/migrations/001_init.sql
psql "postgres://postgres:your_password@localhost:5432/payment_db" -f payment-service/migrations/001_init.sql
psql "postgres://postgres:your_password@localhost:5432/payment_db" -f payment-service/migrations/002_payments.sql
psql "postgres://postgres:your_password@localhost:5432/payment_db" -f payment-service/migrations/003_refunds.sql
```

Replace `your_password` with your PostgreSQL password.

### Redis & Kafka

Start Redis:

```bash
docker run -d --name redis -p 6379:6379 redis:latest
```

Start Kafka:

```bash
docker run -d --name kafka -p 9092:9092 apache/kafka:latest
```

Create the Kafka topic:

```bash
docker exec kafka /opt/kafka/bin/kafka-topics.sh --create --topic payment.events --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
```

Expected local addresses:

```text
Redis: localhost:6379
Kafka: localhost:9092
Kafka topic: payment.events
```

### Run Services

Run each service in a separate terminal:

```bash
cd auth-service
go mod download
go run ./cmd/server
```

```bash
cd payment-service
go mod download
go run ./cmd/server
```

```bash
cd notification-service
go mod download
go run ./cmd/server
```

```bash
cd api-gateway
go mod download
go run ./cmd/server
```

The API Gateway is available at:

```text
http://localhost:8080
```

## Security

* Passwords are bcrypt-hashed before storage.
* JWTs expire after 8 hours.
* There are currently no refresh tokens or server-side token revocation.
* **The API Gateway is intended to be the only publicly exposed application service. Auth and payment services should be deployed on a private network and accessed through the gateway rather than directly from the Internet.**
* **Trust boundary:** the gateway is the only component that verifies JWTs; `payment-service` trusts `X-Merchant-ID`. Direct access to `payment-service` can therefore bypass gateway authentication.
* Customer, payment, and refund queries are scoped by `merchant_id`.
* Idempotency keys prevent duplicate payment creation.
* Rate limiting is implemented per merchant and route using Redis.
* Authentication endpoints are currently not rate-limited.

## Known Limitations

* Payment-provider integration is simulated.
* Development simulation endpoints bypass gateway authentication.
* `payment-service` trusts `X-Merchant-ID`.
* No refresh tokens or token revocation.
* No automated migration tool.
* No Docker Compose setup.
* No automated tests yet.
* Kafka publishing does not use a transactional outbox.
* Notification failures have no retry/DLQ mechanism.
* Service addresses are currently hardcoded.

## Future Improvements

* Real payment-provider integration and verified webhooks
* Transactional outbox for reliable Kafka publishing
* Kafka retries and dead-letter queues
* Docker Compose development environment
* Automated database migrations
* Unit and integration tests
* Structured logging, metrics, and distributed tracing
* Refresh tokens and token revocation
* Environment-based service configuration
* Payment reconciliation

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
