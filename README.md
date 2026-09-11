# Payment Platform

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge\&logo=go\&logoColor=white)
![Gin](https://img.shields.io/badge/Gin-Framework-00ADD8?style=for-the-badge\&logo=gin\&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-4169E1?style=for-the-badge\&logo=postgresql\&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-Rate%20Limiting-DC382D?style=for-the-badge\&logo=redis\&logoColor=white)
![Kafka](https://img.shields.io/badge/Kafka-Event%20Streaming-231F20?style=for-the-badge\&logo=apachekafka\&logoColor=white)
![JWT](https://img.shields.io/badge/JWT-Authentication-000000?style=for-the-badge\&logo=jsonwebtokens\&logoColor=white)

A Go microservices payment platform: an API gateway authenticates and rate-limits requests, an auth service issues JWTs, a payment service owns customers, payments, and refunds, and a notification service asynchronously sends email notifications based on payment events published to Kafka.

This is a backend simulation for learning and portfolio purposes. It does not connect to a real payment processor or bank. Payment success/failure is currently triggered by development-only endpoints that stand in for a real payment provider callback.

## Features

* Registration and login with bcrypt-hashed passwords and JWT authentication
* API gateway as the intended public entry point
* JWT verification at the gateway with downstream identity forwarded via `X-Merchant-ID`
* Redis-backed rate limiting scoped per merchant and route
* Idempotent payment creation using a required `Idempotency-Key`
* Merchant-scoped customer, payment, and refund management
* Payment state machine: `pending → processing → succeeded/failed`
* Full and partial refunds with cumulative refunded-amount tracking
* Kafka-based asynchronous payment notifications
* Event-specific email notifications through SMTP
* Layered handler → service → repository architecture
* UUID identifiers and request validation
* Separate PostgreSQL database ownership for auth and payment data

## Tech Stack

* **Language:** Go 1.26
* **Framework:** [Gin](https://github.com/gin-gonic/gin)
* **Database:** PostgreSQL via [pgx/v5](https://github.com/jackc/pgx)
* **Cache / rate limiting:** Redis via [go-redis/v9](https://github.com/redis/go-redis)
* **Messaging:** Kafka via [segmentio/kafka-go](https://github.com/segmentio/kafka-go)
* **Authentication:** [golang-jwt/jwt](https://github.com/golang-jwt/jwt) + bcrypt
* **Email:** Go `net/smtp`
* **Architecture:** independent Go modules per service, API gateway + event-driven notifications

## Architecture

```mermaid
flowchart LR
    Client[Client]

    subgraph Gateway["api-gateway :8080"]
        MW["JWT auth → Idempotency-Key check → Rate limiter"]
    end

    Auth["auth-service :8081<br/>register, login"]
    Payment["payment-service :8082<br/>customers, payments, refunds"]
    Notification["notification-service<br/>Kafka consumer, no HTTP"]

    Redis[(Redis<br/>rate-limit counters)]
    AuthDB[(PostgreSQL<br/>auth_db)]
    PaymentDB[(PostgreSQL<br/>payment_db)]
    Kafka[/Kafka<br/>payment.events/]
    SMTP[SMTP Server]

    Client --> MW
    MW -->|/auth/register, /auth/login| Auth
    MW -->|/customers, /payments, X-Merchant-ID| Payment
    MW <--> Redis

    Auth --> AuthDB
    Payment --> PaymentDB
    Payment -->|publish| Kafka
    Kafka -->|consume| Notification
    Notification --> SMTP
```

`api-gateway` is the only component that verifies JWTs. After successful verification, it sets `X-Merchant-ID` on the forwarded request; `payment-service` trusts that header instead of verifying the token itself.

This creates a deliberate trust boundary: anything that can reach `payment-service` directly can bypass gateway authentication and supply its own `X-Merchant-ID`. In a real deployment, `payment-service` should not be publicly reachable. See [Security](#security).

## Services

| Service                |   Port | Responsibility                                                                                                 |
| ---------------------- | -----: | -------------------------------------------------------------------------------------------------------------- |
| `api-gateway`          | `8080` | Intended public entry point; JWT verification, idempotency-key validation, rate limiting, and request proxying |
| `auth-service`         | `8081` | Registration, login, password hashing, and JWT issuance                                                        |
| `payment-service`      | `8082` | Customers, payments, payment processing, refunds, and payment event publishing                                 |
| `notification-service` |    `—` | Kafka consumer that sends payment notifications over SMTP; no HTTP server                                      |

Each service is an independent Go module with its own `go.mod`.

## Project Structure

```text
Payment-Platform/
├── .gitignore
├── LICENSE
├── README.md
│
├── api-gateway/
│   ├── .env.example
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── server/
│   │       └── main.go                    entrypoint — router, proxy targets, Redis client
│   └── internal/
│       ├── config/
│       │   └── config.go                  env-based config loader (JWT_SECRET)
│       ├── middlewares/
│       │   ├── auth.go                    JWT verification, sets X-Merchant-ID
│       │   ├── idempotency_key.go         requires + validates Idempotency-Key
│       │   └── rate_limiter.go            Redis-backed per-merchant rate limiter
│       └── proxy/
│           └── proxy.go                   reverse proxy to auth-service / payment-service
│
├── auth-service/
│   ├── .env.example
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── server/
│   │       └── main.go                    entrypoint
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go                  env-based config loader
│   │   ├── db/
│   │   │   └── postgres.go                pgx connection pool
│   │   ├── handlers/
│   │   │   └── auth_handler.go            register / login HTTP handlers
│   │   ├── jwt/
│   │   │   └── jwt.go                     JWT generation (8h expiry)
│   │   ├── models/
│   │   │   └── user_model.go              request/response/domain structs
│   │   ├── password/
│   │   │   └── password.go                bcrypt hash + compare
│   │   ├── repository/
│   │   │   └── user_repository.go         PostgreSQL queries
│   │   ├── services/
│   │   │   └── user_service.go            registration/login business logic
│   │   └── validation/
│   │       └── validation.go              request validation
│   └── migrations/
│       └── 001_init.sql                   users
│
├── payment-service/
│   ├── .env.example
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── server/
│   │       └── main.go                    entrypoint
│   ├── internal/
│   │   ├── config/
│   │   │   └── config.go                  env-based config loader
│   │   ├── db/
│   │   │   └── postgres.go                pgx connection pool
│   │   ├── handlers/
│   │   │   ├── customer_handler.go
│   │   │   ├── payment_handler.go
│   │   │   └── refund_handler.go
│   │   ├── kafka/
│   │   │   └── producer.go                Kafka producer + PaymentEvent schema
│   │   ├── models/
│   │   │   ├── customer_model.go
│   │   │   ├── payment_model.go
│   │   │   └── refund_model.go
│   │   ├── repository/
│   │   │   ├── customer_repository.go
│   │   │   ├── payment_repository.go
│   │   │   └── refund_repository.go
│   │   ├── services/
│   │   │   ├── customer_service.go
│   │   │   ├── payment_service.go         payment lifecycle + event publishing
│   │   │   └── refund_service.go          refund eligibility + amount tracking
│   │   ├── utils/
│   │   │   └── auth.go                    X-Merchant-ID extraction
│   │   └── validations/
│   │       ├── customer_validation.go
│   │       └── payment_validation.go      currency + payment transition rules
│   └── migrations/
│       ├── 001_init.sql                   customers
│       ├── 002_payments.sql               payments
│       └── 003_refunds.sql                refunds
│
└── notification-service/
    ├── .env.example
    ├── go.mod
    ├── go.sum
    ├── cmd/
    │   └── server/
    │       └── main.go                    Kafka consumer wiring, no HTTP server
    └── internal/
        ├── config/
        │   └── config.go                  SMTP configuration
        ├── kafka/
        │   └── consumer.go                Kafka consumer
        ├── mail/
        │   └── sender.go                  SMTP sender
        ├── models/
        │   └── payment_model.go           PaymentEvent struct
        └── services/
            └── payment_notification_service.go
```

## Authentication

> **Naming note:** `auth-service` stores a `User` in the `users` table. `payment-service` uses the authenticated user's ID as `merchant_id`. There is no separate merchant entity; "merchant" is the payment-domain name for the authenticated user.

### Registration

```text
Client → api-gateway → auth-service
                          ├── validate request
                          ├── hash password with bcrypt
                          ├── insert user
                          └── 201 Created
```

### Login

```text
Client → api-gateway → auth-service
                          ├── look up user by email
                          ├── compare bcrypt hash
                          ├── generate JWT
                          └── 200 OK { token }
```

The JWT contains the user's ID in the `id` claim and expires after 8 hours. There is currently no refresh token or server-side revocation.

## Payment Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant G as api-gateway
    participant P as payment-service
    participant Pg as Postgres
    participant K as Kafka
    participant N as notification-service

    C->>G: POST /payments<br/>Bearer JWT + Idempotency-Key
    G->>G: verify JWT → set X-Merchant-ID
    G->>G: validate Idempotency-Key
    G->>G: rate limit by merchant + route
    G->>P: proxy request

    P->>Pg: verify customer belongs to merchant
    P->>Pg: check idempotency key

    alt existing idempotency key
        P-->>C: 201 Created (existing payment)
    else new payment
        P->>Pg: insert payment (pending)
        P->>K: publish payment.created
        P-->>C: 201 Created
        K->>N: consume event
        N->>N: send "Payment Created" email
    end

    C->>G: POST /payments/:id/process
    G->>P: proxy authenticated request
    P->>Pg: validate pending → processing
    P->>Pg: update status → processing
    P->>K: publish payment.processing
    P-->>C: 200 OK
    K->>N: consume event
    N->>N: send "Payment Processing" email

    Note over C,P: Development-only simulation<br/>stands in for a provider callback

    C->>P: POST /payments/:id/succeed or /fail<br/>X-Merchant-ID required
    P->>Pg: validate processing → succeeded/failed
    P->>K: publish payment.succeeded/failed
    K->>N: consume event
    N->>N: send outcome email
```

### Idempotency

```http
POST /payments
Authorization: Bearer <JWT>
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
```

The `Idempotency-Key` is required and validated at both the gateway and `payment-service`.

The database enforces:

```text
UNIQUE(merchant_id, idempotency_key)
```

If the same merchant retries `POST /payments` with the same key, the existing payment is returned instead of creating another payment.

This protects against duplicate payment creation when a client retries after a network failure.

## Payment State Machine

The core payment lifecycle is:

```mermaid
stateDiagram-v2
    [*] --> pending
    pending --> processing
    processing --> succeeded
    processing --> failed
```

The service rejects invalid lifecycle transitions such as:

```text
pending → succeeded
pending → failed
succeeded → processing
failed → processing
```

Refund statuses are handled separately:

```text
succeeded
    ↓
partially_refunded
    ↓
refunded
```

## Refunds

```mermaid
flowchart TD
    Payment["Payment (succeeded or partially_refunded)"]
    Payment --> Check{"refund amount <=<br/>remaining refundable amount?"}

    Check -- no --> Reject["Reject refund"]
    Check -- yes --> Insert["Insert refund + update cumulative refunded amount"]

    Insert --> Full{"Fully refunded?"}
    Full -- yes --> Refunded["status → refunded<br/>publish payment.refunded"]
    Full -- no --> Partial["status → partially_refunded<br/>publish payment.partially_refunded"]
```

Refunds are allowed only when the payment is `succeeded` or `partially_refunded`.

The service ensures:

```text
refund amount <= amount_cents - refunded_amount_cents
```

Multiple partial refunds can be issued until the payment is fully refunded.

The refund record and payment's cumulative refunded amount are updated in the same PostgreSQL transaction.

## Payment Events

`payment-service` publishes the following events to the Kafka topic `payment.events`:

* `payment.created`
* `payment.processing`
* `payment.succeeded`
* `payment.failed`
* `payment.refunded`
* `payment.partially_refunded`

`notification-service` consumes these events using the `notification-service` consumer group and sends an event-specific email through SMTP.

Notifications are intentionally outside the payment request path. A slow or unavailable SMTP server therefore does not directly block the payment HTTP request.

> **Current limitation:** notification send failures are logged but are not retried or placed in a dead-letter queue yet. Kafka retry/DLQ handling is planned as a future improvement.

## API Endpoints

All normal API routes below are exposed through:

```text
http://localhost:8080
```

The gateway applies rate limits based on `merchant_id` + route.

### Authentication

| Method | Path             | Auth | Rate Limit | Notes                                |
| ------ | ---------------- | ---- | ---------- | ------------------------------------ |
| POST   | `/auth/register` | —    | —          | Proxied to auth-service              |
| POST   | `/auth/login`    | —    | —          | Proxied to auth-service; returns JWT |

**POST /auth/register**

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "at-least-8-chars"
}
```

**POST /auth/login**

```json
{
  "email": "jane@example.com",
  "password": "your-password"
}
```

Returns:

```json
{
  "token": "<jwt>"
}
```

### Customers

| Method | Path             | Auth | Rate Limit |
| ------ | ---------------- | ---- | ---------: |
| POST   | `/customers`     | JWT  |   20 / min |
| GET    | `/customers`     | JWT  |   60 / min |
| GET    | `/customers/:id` | JWT  |   60 / min |

**POST /customers**

```json
{
  "name": "John Smith",
  "email": "john@example.com"
}
```

### Payments

| Method | Path                    | Auth                    | Rate Limit | Notes                   |
| ------ | ----------------------- | ----------------------- | ---------: | ----------------------- |
| POST   | `/payments`             | JWT + `Idempotency-Key` |   20 / min | Creates pending payment |
| GET    | `/payments`             | JWT                     |   60 / min |                         |
| GET    | `/payments/:id`         | JWT                     |   60 / min |                         |
| POST   | `/payments/:id/process` | JWT                     |   20 / min | `pending → processing`  |

**POST /payments**

Headers:

```http
Authorization: Bearer <JWT>
Idempotency-Key: <uuid>
```

Body:

```json
{
  "customer_id": "<uuid>",
  "amount_cents": 5000,
  "currency": "USD"
}
```

Supported currencies:

```text
USD
EUR
TRY
```

### Refunds

| Method | Path                    | Auth | Rate Limit | Notes                  |
| ------ | ----------------------- | ---- | ---------: | ---------------------- |
| POST   | `/payments/:id/refunds` | JWT  |   10 / min | Full or partial refund |
| GET    | `/payments/:id/refunds` | JWT  |   60 / min |                        |

**POST /payments/:id/refunds**

```json
{
  "amount_cents": 2000
}
```

### Development / Testing Only

These endpoints are registered directly on `payment-service` and therefore bypass:

* JWT verification by the gateway
* gateway rate limiting
* gateway `X-Merchant-ID` injection

However, **they still require a valid `X-Merchant-ID` header supplied directly by the caller**, because the payment handlers use that header to determine the merchant.

```http
X-Merchant-ID: <merchant-uuid>
```

| Method | Path                    | Notes                                    |
| ------ | ----------------------- | ---------------------------------------- |
| POST   | `/payments/:id/succeed` | Simulates a successful provider callback |
| POST   | `/payments/:id/fail`    | Simulates a failed provider callback     |

These endpoints are development-only stand-ins for a real payment provider integration and should be removed when real provider callbacks/webhooks are implemented.

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

    CUSTOMERS ||--o{ PAYMENTS : "has"
    PAYMENTS ||--o{ REFUNDS : "has"
```

`USERS` lives in `auth_db`, owned by `auth-service`.

`CUSTOMERS`, `PAYMENTS`, and `REFUNDS` live in `payment_db`, owned by `payment-service`.

There is no foreign key between `users` and the payment tables. `merchant_id` is the authenticated user's ID carried across the service boundary via `X-Merchant-ID`.

The database enforces:

```text
users.email                          UNIQUE
customers.(merchant_id, email)       UNIQUE
payments.(merchant_id, idempotency_key) UNIQUE
```

`payments.customer_id` references `customers.id`, and `refunds.payment_id` references `payments.id`.

## Setup

### Prerequisites

* Go 1.26+
* PostgreSQL
* Docker
* Redis
* Kafka
* SMTP account/relay for outgoing email

### Clone

```bash
git clone https://github.com/ErenKarakus1/Payment-Platform.git
cd Payment-Platform
```

### Environment Variables

Each service has its own `.env.example`.

Copy each example to `.env`:

```bash
cp api-gateway/.env.example api-gateway/.env
cp auth-service/.env.example auth-service/.env
cp payment-service/.env.example payment-service/.env
cp notification-service/.env.example notification-service/.env
```

Then fill in the values for your local environment.

The services currently require their `.env` files at startup.

> **Important:** `JWT_SECRET` must be identical in `api-gateway` and `auth-service`, because the gateway verifies JWTs issued by the auth service.

If using Gmail SMTP, use a Google App Password rather than your normal account password.

Never commit real `.env` files or credentials.

### Database

Create the two PostgreSQL databases:

```sql
CREATE DATABASE auth_db;
CREATE DATABASE payment_db;
```

Apply the migrations manually:

```bash
psql "postgres://postgres:your_password@localhost:5432/auth_db" \
  -f auth-service/migrations/001_init.sql

psql "postgres://postgres:your_password@localhost:5432/payment_db" \
  -f payment-service/migrations/001_init.sql

psql "postgres://postgres:your_password@localhost:5432/payment_db" \
  -f payment-service/migrations/002_payments.sql

psql "postgres://postgres:your_password@localhost:5432/payment_db" \
  -f payment-service/migrations/003_refunds.sql
```

Replace `your_password` with your local PostgreSQL password.

There is currently no migration tool wired into the project, so migrations are applied manually.

### Redis & Kafka

Start Redis:

```bash
docker run -d --name redis -p 6379:6379 redis:latest
```

Start Kafka:

```bash
docker run -d --name kafka -p 9092:9092 apache/kafka:latest
```

Create the payment events topic:

```bash
docker exec kafka /opt/kafka/bin/kafka-topics.sh \
  --create \
  --topic payment.events \
  --bootstrap-server localhost:9092 \
  --partitions 1 \
  --replication-factor 1
```

The application currently expects:

```text
Redis: localhost:6379
Kafka: localhost:9092
Kafka topic: payment.events
```

The gateway also currently expects:

```text
auth-service:    http://localhost:8081
payment-service: http://localhost:8082
```

These addresses are hardcoded in the application rather than environment-configurable. If you change the deployment topology, the source code currently needs to be updated.

### Run

Start PostgreSQL, Redis, and Kafka first.

Then start each application service in a **separate terminal**.

**Terminal 1 — auth-service**

```bash
cd auth-service
go run ./cmd/server
```

Runs on `:8081`.

**Terminal 2 — payment-service**

```bash
cd payment-service
go run ./cmd/server
```

Runs on `:8082`.

**Terminal 3 — notification-service**

```bash
cd notification-service
go run ./cmd/server
```

Kafka consumer only; no HTTP port.

**Terminal 4 — api-gateway**

```bash
cd api-gateway
go run ./cmd/server
```

Runs on `:8080` and is the intended public entry point.

Each service is a separate Go module. If dependencies have not already been downloaded, run:

```bash
go mod download
```

inside the corresponding service directory.

## Security

* Passwords are bcrypt-hashed before storage; plaintext passwords are never persisted.
* JWTs use an 8-hour expiration and contain the authenticated user's ID in the `id` claim.
* There are currently no refresh tokens or server-side token revocation mechanisms.
* **The API Gateway is intended to be the only publicly exposed application service. Auth and payment services should be deployed on a private network and accessed through the gateway rather than directly from the Internet.**
* **Trust boundary:** `api-gateway` is the only component that verifies JWTs. `payment-service` trusts the `X-Merchant-ID` header and does not independently verify the token.
* `payment-service` should not be publicly exposed. A caller that can reach it directly can bypass gateway authentication and supply its own `X-Merchant-ID`.
* Customer, payment, and refund queries are scoped by `merchant_id` in the payment service.
* Idempotency keys prevent duplicate payment creation when clients retry the same request.
* Rate limiting is implemented per merchant and per route using Redis.
* There is currently no separate per-IP rate-limiting layer.
* Authentication endpoints themselves are not currently rate-limited by the gateway.

## Design Decisions

### API Gateway as the Intended Single Entry Point

Clients are intended to access `auth-service` and `payment-service` through the gateway rather than depending on the internal service topology.

The gateway provides a centralized location for:

* JWT verification
* identity propagation
* idempotency-key validation
* rate limiting
* request proxying

### Kafka for Notifications

Email is not sent synchronously from the payment request.

Instead:

```text
payment-service
      ↓
    Kafka
      ↓
notification-service
      ↓
     SMTP
```

This keeps SMTP latency and failures out of the payment API request path.

### Separate Databases per Service

`auth-service` and `payment-service` own separate PostgreSQL databases:

```text
auth-service    → auth_db
payment-service → payment_db
```

There are no shared tables or cross-service foreign keys.

The services are connected at the domain level through:

```text
merchant_id = authenticated user's ID
```

This preserves service-level database ownership at the cost of not having database-enforced referential integrity between users and payment records.

### Idempotency Keys

Payment creation requires a client-generated idempotency key.

This allows a client to safely retry a payment creation request after a network failure without creating another payment record.

## Known Limitations

* Redis, Kafka, and gateway proxy addresses are hardcoded.
* `/payments/:id/succeed` and `/payments/:id/fail` are development-only provider simulations.
* Development simulation endpoints bypass gateway authentication and rate limiting and rely on a caller-supplied `X-Merchant-ID`.
* No refresh tokens or server-side JWT revocation.
* `payment-service` trusts `X-Merchant-ID` from callers that can reach it directly.
* Authentication routes have no gateway rate limiting.
* No authenticated `whoami` / profile endpoint.
* No Docker Compose setup.
* Database migrations are applied manually.
* No automated tests.
* Kafka publishing is not protected by a transactional outbox, so database state changes and event publication are not atomic.
* Notification send failures are logged without retry or dead-letter handling.

## Future Improvements

* Replace hardcoded Redis, Kafka, and proxy addresses with environment-based configuration
* Add Docker Compose for the complete local environment
* Add refresh tokens and server-side token revocation
* Replace development payment simulations with a real provider integration and verified webhooks
* Independently verify identity in `payment-service`, or network-isolate it so only the gateway can reach it
* Add automated database migrations
* Add Kafka retry and dead-letter handling
* Add a transactional outbox for reliable payment event publication
* Add structured logging, distributed tracing, and metrics
* Add unit and integration tests for PostgreSQL, Redis, Kafka, and service boundaries
* Add payment reconciliation
* Add health and readiness endpoints

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
