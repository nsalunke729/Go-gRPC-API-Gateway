// Package server is a public facade over the internal gateway wiring.
// It exists so that the Vercel api/index.go handler (built by Vercel as
// "handler/api", outside the module path) can import gateway setup without
// hitting Go's internal-package visibility restriction.
package server

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/nsalunke729/go-grpc-gateway/internal/embed"
	"github.com/nsalunke729/go-grpc-gateway/internal/gateway"
	"github.com/nsalunke729/go-grpc-gateway/internal/ordersvc"
	"github.com/nsalunke729/go-grpc-gateway/internal/usersvc"
)

// NewHandler returns a fully wired http.Handler using in-process service adapters.
// Suitable for serverless deployments where gRPC connections between processes
// are not viable.
func NewHandler(jwtSecret string, rateLimit float64, rateBurst int) http.Handler {
	log, _ := zap.NewProduction()

	cfg := gateway.Config{
		Addr:      ":8080",
		JWTSecret: jwtSecret,
		RateLimit: rateLimit,
		RateBurst: rateBurst,
	}

	userClient := embed.NewUserAdapter(usersvc.NewServer())
	orderClient := embed.NewOrderAdapter(ordersvc.NewServer())

	return gateway.NewWithClients(cfg, userClient, orderClient, log).HTTPHandler()
}
