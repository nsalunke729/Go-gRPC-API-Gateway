# syntax=docker/dockerfile:1

# Multi-stage build. Set CMD build arg to the binary to compile:
#   docker build --build-arg CMD=gateway -t gateway .
#   docker build --build-arg CMD=user-svc -t user-svc .
#   docker build --build-arg CMD=order-svc -t order-svc .

FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG CMD=gateway
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/service ./cmd/${CMD}

FROM scratch
COPY --from=builder /bin/service /bin/service
ENTRYPOINT ["/bin/service"]
