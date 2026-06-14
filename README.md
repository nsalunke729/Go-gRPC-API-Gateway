# Go gRPC API Gateway

A lightweight API gateway written in Go that exposes a REST API to external clients and translates requests into gRPC calls to backend microservices.

**Live demo:** [https://go-grpc-api-gateway.vercel.app](https://go-grpc-api-gateway.vercel.app/healthz)  
_(replace with your Vercel URL — see [Vercel deployment](#vercel-deployment) section)_

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

> On Vercel (serverless) all three services run in-process via the `server/` facade and `internal/embed` adapters — same REST API and middleware, no separate gRPC network hop needed.

## Features

- **REST → gRPC translation** — gateway accepts JSON over HTTP and fans out to typed gRPC calls
- **JWT authentication** — HS256 Bearer token validation on all non-health endpoints
- **Per-client rate limiting** — token-bucket limiter (100 req/s, burst 200) keyed by `RemoteAddr`
- **Structured logging** — every request logged with method, path, status, and latency via [zap](https://github.com/uber-go/zap)
- **Graceful shutdown** — SIGINT/SIGTERM drains in-flight requests before exit
- **Health check** — `GET /healthz` with no auth required
- **Vercel serverless** — `api/index.go` entry point, deploys in one click
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
| Serverless | [Vercel Go runtime](https://vercel.com/docs/functions/runtimes/go) |

> **Codec note:** The project overrides the default gRPC protobuf codec with a JSON codec (`internal/codec/json.go`). The `.proto` files in `proto/` document the service contracts but no `protoc` / `buf generate` step is needed to build. Swapping in real protobuf is a one-line change to the codec.

## Project structure

```
api/
  index.go      Vercel serverless entrypoint (imports server/ only)

cmd/
  gateway/      HTTP gateway entrypoint (standalone mode)
  user-svc/     User gRPC service entrypoint
  order-svc/    Order gRPC service entrypoint
  gentoken/     Dev JWT generator

server/
  server.go     Public facade — wires gateway with in-process adapters
                (lets api/index.go avoid Go's internal-package restriction)

internal/
  codec/        JSON-over-gRPC codec (replaces default proto codec)
  embed/        In-process adapters: UserAdapter, OrderAdapter
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
scripts/
  start-local.ps1  One-command local start with log tailing (Windows)
```

## Getting started

### Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) (optional, for `make up`)

### Run locally — one command (Windows)

```powershell
.\scripts\start-local.ps1
# Starts all three services, prints a ready-to-use JWT, tails logs in colour.
# Stop: .\scripts\start-local.ps1 -Stop
```

### Run locally — three terminals

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
# prints a 24-hour HS256 token signed with JWT_SECRET (default: "dev-secret-change-me")
```

## API reference

All endpoints except `/healthz` require `Authorization: Bearer <token>`.

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/healthz` | Health check (no auth) |
| `POST` | `/users` | Create a user |
| `GET` | `/users/{id}` | Get a user by ID |
| `POST` | `/orders` | Create an order |
| `GET` | `/orders/{id}` | Get an order by ID |
| `GET` | `/users/{userID}/orders` | List all orders for a user |

### Quick test (local)

```bash
TOKEN=$(go run ./cmd/gentoken)

# Create user
curl -X POST http://localhost:8080/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
# → {"user":{"id":"usr_...","name":"Alice","email":"alice@example.com","created_at":...}}

# Create order
curl -X POST http://localhost:8080/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"usr_...","amount":49.99}'

# 401 check — missing token
curl http://localhost:8080/users/usr_...
```

### Quick test (Vercel live deployment)

1. Set `JWT_SECRET` to a strong random string in your Vercel project's **Environment Variables**.
2. Generate a matching token locally:
   ```powershell
   $env:JWT_SECRET = "the-secret-you-set-in-vercel"
   $TOKEN = go run ./cmd/gentoken
   ```
3. Use the printed token against your deployment URL.

   > **Windows PowerShell:** use `curl.exe` (real curl binary) not `curl` (which aliases `Invoke-WebRequest` and uses different flags). Omit the trailing slash from `$BASE`.

   ```powershell
   # PowerShell
   $BASE = "https://<your-vercel-url>.vercel.app"   # no trailing slash

   curl.exe "$BASE/healthz"

   curl.exe -X POST "$BASE/users" `
     -H "Authorization: Bearer $TOKEN" `
     -H "Content-Type: application/json" `
     -d '{\"name\":\"Alice\",\"email\":\"alice@example.com\"}'
   ```

   ```bash
   # bash / macOS / Linux / WSL
   BASE="https://<your-vercel-url>.vercel.app"

   curl "$BASE/healthz"

   curl -X POST "$BASE/users" \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"name":"Alice","email":"alice@example.com"}'
   ```

> **Stateless note:** Vercel functions are stateless. Data created in one request may not persist if the function goes cold between requests. For a persistent store, swap the in-memory maps for a database (e.g. Vercel Postgres, PlanetScale, or Upstash Redis).

## Vercel deployment

1. Go to [vercel.com](https://vercel.com) → **Add New Project** → import `nsalunke729/Go-gRPC-API-Gateway`
2. Vercel auto-detects the Go runtime from `api/index.go` — no build settings needed
3. Add one environment variable:

   | Key | Value |
   |-----|-------|
   | `JWT_SECRET` | any strong secret string |

4. Click **Deploy**

Every push to `main` triggers an automatic redeploy.

## Configuration

### Standalone gateway (local / Docker)

| Variable | Default | Description |
|----------|---------|-------------|
| `GATEWAY_ADDR` | `:8080` | Listen address |
| `JWT_SECRET` | `dev-secret-change-me` | HMAC signing secret |
| `USER_SVC_ADDR` | `localhost:9001` | user-svc gRPC address |
| `ORDER_SVC_ADDR` | `localhost:9002` | order-svc gRPC address |
| `RATE_LIMIT` | `100` | Requests/sec per client |
| `RATE_BURST` | `200` | Burst size per client |

### Vercel (serverless)

| Variable | Required | Description |
|----------|----------|-------------|
| `JWT_SECRET` | Yes | HMAC signing secret |
| `RATE_LIMIT` | No (default 100) | Requests/sec per client |
| `RATE_BURST` | No (default 200) | Burst size per client |

## Running tests

```bash
make test
```

## Makefile targets

| Target | Description |
|--------|-------------|
| `make build` | Build all binaries into `bin/` |
| `make test` | Run unit tests |
| `make test-race` | Run with race detector (requires CGO / Linux) |
| `make lint` | Run golangci-lint |
| `make up` | Docker Compose build + start |
| `make down` | Docker Compose stop |
| `make proto` | Regenerate Go from .proto (requires buf) |
