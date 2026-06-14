.PHONY: build test lint up down proto tidy

# Build all three binaries into the repo root
build:
	go build -o bin/gateway    ./cmd/gateway
	go build -o bin/user-svc  ./cmd/user-svc
	go build -o bin/order-svc ./cmd/order-svc

test:
	go test -count=1 ./...

# Race detector requires CGO; CI (Linux) enables it automatically
test-race:
	go test -race -count=1 ./...

lint:
	golangci-lint run ./...

# Regenerate Go code from .proto files (requires buf: https://buf.build)
proto:
	buf generate

# Tidy and verify the module graph
tidy:
	go mod tidy
	go mod verify

# Start all services with Docker Compose
up:
	docker compose up --build

down:
	docker compose down

# Run services locally (three separate terminals) — for development
run-user-svc:
	go run ./cmd/user-svc

run-order-svc:
	go run ./cmd/order-svc

run-gateway:
	go run ./cmd/gateway
