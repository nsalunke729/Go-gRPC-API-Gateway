// Package handler is the Vercel serverless entry point.
// All three services run in-process via embed.UserAdapter / embed.OrderAdapter
// so no gRPC network connections are needed.
package handler

import (
	"net/http"
	"os"
	"strconv"
	"sync"

	"go.uber.org/zap"

	"github.com/nsalunke729/go-grpc-gateway/internal/embed"
	"github.com/nsalunke729/go-grpc-gateway/internal/gateway"
	"github.com/nsalunke729/go-grpc-gateway/internal/ordersvc"
	"github.com/nsalunke729/go-grpc-gateway/internal/usersvc"
)

var (
	once    sync.Once
	handler http.Handler
)

// Handler is the Vercel-compatible entrypoint. Vercel calls this for every request.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		log, _ := zap.NewProduction()

		rateLimit, _ := strconv.ParseFloat(getenv("RATE_LIMIT", "100"), 64)
		rateBurst, _ := strconv.Atoi(getenv("RATE_BURST", "200"))

		cfg := gateway.Config{
			Addr:      ":8080", // unused in serverless — Vercel manages the port
			JWTSecret: getenv("JWT_SECRET", "dev-secret-change-me"),
			RateLimit: rateLimit,
			RateBurst: rateBurst,
		}

		userClient := embed.NewUserAdapter(usersvc.NewServer())
		orderClient := embed.NewOrderAdapter(ordersvc.NewServer())

		srv := gateway.NewWithClients(cfg, userClient, orderClient, log)
		handler = srv.HTTPHandler()
	})

	handler.ServeHTTP(w, r)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
