# Go gRPC API Gateway

A lightweight API gateway written in Go that exposes a REST API to external clients and translates requests into gRPC calls to backend microservices.

## Architecture

```
Client (REST/JSON)
       │
       ▼
┌─────────────────────────────────┐
│           API Gateway           │  :8080
│  ┌──────────┐  ┌─────────────┐  │
│  │ JWT Auth │  │ Rate Limiter│  │  middleware
│  └──────────┘  └─────────────┘  │
│  ┌──────────┐  ┌─────────────┐  │
│  │  /users  │  │  /orders    │  │  handlers
│  └────┬─────┘  └──────┬──────┘  │
└───────┼────────────────┼─────────┘
        │   gRPC (JSON)  │
   ┌────▼────┐      ┌────▼─────┐
   │user-svc │      │order-svc │
   │  :9001  │      │  :9002   │
   └─────────┘      └──────────┘
```

## Features

- **REST → gRPC translation** — gateway accepts JSON over HTTP and fans out to typed gRPC calls
- **JWT authentication** — HS256 Bearer token validation on all non-health endpoints
- **Per-client rate limiting** — token-bucket limiter (100 req/s, burst 200) keyed by `RemoteAddr`
- **Structured logging** — every request logged with method, path, status, and latency via [zap](https://github.com/uber-go/zap)
- **Graceful shutdown** — SIGINT/SIGTERM drains in-flight requests before exit
- **Health check** — `GET /healthz` with no auth required
- **Docker + Compose** — single multi-stage `Dockerfile`, `docker compose up` starts everything
- **CI** — GitHub Actions: vet, test, build all binaries, golangci-lint

## Tech stack

| Layer | Choice |
|---|---|
| Language | Go 1.26 |
| HTTP router | [chi v5](https://github.com/go-chi/chi) |
| gRPC | [google.golang.org/grpc](https://pkg.go.dev/google.golang.org/grpc) |
| JWT | [golang-jwt/jwt v5](https://github.com/golang-jwt/jwt) |
| Rate limiting | [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) |
| Logging | [uber-go/zap](https://github.com/uber-go/zap) |

> **Codec note:** The project overrides the default gRPC protobuf codec with a JSON codec (`internal/codec/json.go`). This means the `.proto` files in `proto/` document the service contracts but no `protoc` / `buf generate` step is needed to build. Swapping in real protobuf is a one-line change to the codec.

## Project structure

```
cmd/
  gateway/      HTTP gateway entrypoint
  user-svc/     User gRPC service entrypoint
  order-svc/    Order gRPC service entrypoint
  gentoken/     Dev JWT generator

internal/
  codec/        JSON-over-gRPC codec (replaces default proto codec)
  pb/
    user/       Request/response types for UserService
    order/      Request/response types for OrderService
  usersvc/      gRPC server impl, service descriptor, client
  ordersvc/     gRPC server impl, service descriptor, client
  gateway/
    middleware/ JWT auth + rate limiter
    handlers/   REST handlers (translate to gRPC calls)
    server.go   chi router wiring

proto/          .proto files (contract documentation)
```

## Getting started

### Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) (optional, for `make up`)

### Run locally (three terminals)

```bash
go run ./cmd/user-svc     # terminal 1 — listens on :9001
go run ./cmd/order-svc    # terminal 2 — listens on :9002
go run ./cmd/gateway      # terminal 3 — listens on :8080
```

### Run with Docker Compose

```bash
make up
```

### Generate a dev JWT

```bash
go run ./cmd/gentoken
# prints a signed token using the default secret "dev-secret-change-me"
```

## API reference

All endpoints except `/healthz` require `Authorization: Bearer <token>`.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/healthz` | Health check |
| `POST` | `/users` | Create a user |
| `GET` | `/users/{id}` | Get a user by ID |
| `POST` | `/orders` | Create an order |
| `GET` | `/orders/{id}` | Get an order by ID |
| `GET` | `/users/{userID}/orders` | List all orders for a user |

### Example

```bash
TOKEN=$(go run ./cmd/gentoken)

# Create user
curl -X POST http://localhost:8080/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'

# {"user":{"id":"usr_1234...","name":"Alice","email":"alice@example.com","created_at":1234567890}}

# Create order
curl -X POST http://localhost:8080/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"usr_1234...","amount":49.99}'
```

## Configuration

All services are configured via environment variables.

### Gateway

| Variable | Default | Description |
|----------|---------|-------------|
| `GATEWAY_ADDR` | `:8080` | Listen address |
| `JWT_SECRET` | `dev-secret-change-me` | HMAC signing secret |
| `USER_SVC_ADDR` | `localhost:9001` | user-svc gRPC address |
| `ORDER_SVC_ADDR` | `localhost:9002` | order-svc gRPC address |
| `RATE_LIMIT` | `100` | Requests/sec per client |
| `RATE_BURST` | `200` | Burst size per client |

### user-svc / order-svc

| Variable | Default | Description |
|----------|---------|-------------|
| `USER_SVC_ADDR` | `:9001` | Listen address |
| `ORDER_SVC_ADDR` | `:9002` | Listen address |

## Running tests

```bash
make test
```

## Makefile targets

| Target | Description |
|--------|-------------|
| `make build` | Build all binaries into `bin/` |
| `make test` | Run unit tests |
| `make lint` | Run golangci-lint |
| `make up` | Docker Compose build + start |
| `make down` | Docker Compose stop |
| `make proto` | Regenerate Go from .proto (requires buf) |
